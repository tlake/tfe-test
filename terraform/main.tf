variable "aws_access_key" {
  type = "string"
}

variable "aws_secret_key" {
  type = "string"
}

variable "aws_region" {
  type = "string"
}

variable "vpc_name" {
  type = "string"
}

variable "vpc_cidr_block" {
  type = "string"
}

variable "ssh_key_name" {
  type = "string"
}

variable "ssh_key_contents" {
  type = "string"
}

output "elb_dns_name" {
  value = "${aws_elb.elb.dns_name}"
}

provider "aws" {
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  region     = "${var.aws_region}"
}

data "aws_availability_zones" "zones" {}

data "aws_iam_policy_document" "flow_log_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["vpc-flow-logs.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "flow_log" {
  statement {
    effect    = "Allow"
    resources = ["*"]

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
    ]
  }
}

data "aws_ami" "amazon_linux" {
  most_recent = true

  filter {
    name = "name"

    values = [
      "amzn-ami-hvm-*-x86_64-gp2",
    ]
  }

  filter {
    name = "owner-alias"

    values = [
      "amazon",
    ]
  }
}

data "aws_acm_certificate" "neworca" {
  domain   = "*.neworca.io"
  statuses = ["ISSUED"]
}

################################
# VPC
################################

resource "aws_vpc" "vpc" {
  cidr_block = "${var.vpc_cidr_block}"

  enable_dns_hostnames = true
  enable_dns_support   = true

  tags {
    Name = "${format("%s", var.vpc_name)}"
  }
}

locals {
  public_subnets = [
    "${cidrsubnet(var.vpc_cidr_block, 3, 0)}",
  ]

  private_subnets = [
    "${cidrsubnet(var.vpc_cidr_block, 3, 7)}",
  ]
}

resource "aws_subnet" "public" {
  count = "${length(local.public_subnets)}"

  vpc_id                  = "${aws_vpc.vpc.id}"
  cidr_block              = "${element(local.public_subnets, count.index)}"
  availability_zone       = "${element(data.aws_availability_zones.zones.names, count.index)}"
  map_public_ip_on_launch = true

  tags {
    Name = "${format("%s Public %d", var.vpc_name, count.index + 1)}"
  }
}

resource "aws_subnet" "private" {
  count = "${length(local.private_subnets)}"

  vpc_id                  = "${aws_vpc.vpc.id}"
  cidr_block              = "${element(local.private_subnets, count.index)}"
  availability_zone       = "${element(data.aws_availability_zones.zones.names, count.index)}"
  map_public_ip_on_launch = false

  tags {
    Name = "${format("%s Private %d", var.vpc_name, count.index + 1)}"
  }
}

resource "aws_vpc_endpoint" "private_s3" {
  vpc_id       = "${aws_vpc.vpc.id}"
  service_name = "com.amazonaws.${var.aws_region}.s3"

  route_table_ids = [
    "${aws_route_table.private.*.id}",
    "${aws_route_table.public.*.id}",
  ]
}

resource "aws_route" "nat_routes" {
  count                  = "${length(local.private_subnets)}"
  destination_cidr_block = "0.0.0.0/0"

  route_table_id = "${element(aws_route_table.private.*.id, count.index)}"
  nat_gateway_id = "${element(aws_nat_gateway.private.*.id, count.index)}"
}

resource "aws_eip" "nat_eip" {
  count = "${length(local.private_subnets)}"
  vpc   = true
}

resource "aws_nat_gateway" "private" {
  count = "${length(local.private_subnets)}"

  allocation_id = "${element(aws_eip.nat_eip.*.id, count.index)}"
  subnet_id     = "${element(aws_subnet.public.*.id, count.index)}"
}

resource "aws_internet_gateway" "vpc" {
  vpc_id = "${aws_vpc.vpc.id}"

  tags {
    Name = "${format("%s Gateway", var.vpc_name)}"
  }
}

resource "aws_route_table" "private" {
  count  = "${length(local.private_subnets)}"
  vpc_id = "${aws_vpc.vpc.id}"

  tags {
    Name = "${format("%s Private", var.vpc_name)}"
  }
}

resource "aws_route_table_association" "private" {
  count = "${length(local.private_subnets)}"

  subnet_id      = "${element(aws_subnet.private.*.id, count.index)}"
  route_table_id = "${element(aws_route_table.private.*.id, count.index)}"
}

resource "aws_route_table" "public" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.vpc.id}"
  }

  tags {
    Name = "${format("%s Public", var.vpc_name)}"
  }
}

resource "aws_route_table_association" "public" {
  count = "${length(local.public_subnets)}"

  subnet_id      = "${element(aws_subnet.public.*.id, count.index)}"
  route_table_id = "${aws_route_table.public.id}"
}

################################
# IAM
################################

resource "aws_iam_role" "vpc_role" {
  name               = "${lower(replace(var.vpc_name, " ", "-"))}-vpc-flow-logs"
  assume_role_policy = "${data.aws_iam_policy_document.flow_log_assume_role.json}"
}

resource "aws_iam_role_policy" "vpc_role_policy" {
  name   = "${lower(replace(var.vpc_name, " ", "-"))}-vpc-flow-logs"
  role   = "${aws_iam_role.vpc_role.id}"
  policy = "${data.aws_iam_policy_document.flow_log.json}"
}

################################
# Logging
################################

resource "aws_cloudwatch_log_group" "vpc_log_group" {
  name = "${lower(replace(var.vpc_name, " ", "-"))}-vpc-flow-logs"
}

resource "aws_flow_log" "vpc_flow_log" {
  log_group_name = "${aws_cloudwatch_log_group.vpc_log_group.name}"
  iam_role_arn   = "${aws_iam_role.vpc_role.arn}"
  vpc_id         = "${aws_vpc.vpc.id}"
  traffic_type   = "ALL"
}

################################
# Security Groups
################################

resource "aws_security_group" "vpc-sg" {
  name        = "${var.vpc_name}-vpc-sg"
  vpc_id      = "${aws_vpc.vpc.id}"
  description = "TFE VPC security group"

  ingress = {
    self        = true
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  egress = {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  tags = {
    Name = "${var.vpc_name}-vpc-sg"
  }
}

resource "aws_security_group" "svc-sg" {
  name        = "${var.vpc_name}-svc-sg"
  vpc_id      = "${aws_vpc.vpc.id}"
  description = "Service security group"

  ingress = {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress = {
    protocol    = "tcp"
    from_port   = 443
    to_port     = 443
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress = {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.vpc_name}-svc-sg"
  }
}

################################
# ELB
################################

resource "aws_elb" "elb" {
  name    = "${var.vpc_name}"
  subnets = ["${aws_subnet.public.id}"]

  security_groups = [
    "${aws_security_group.vpc-sg.id}",
    "${aws_security_group.svc-sg.id}",
  ]

  listener = {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  listener = {
    instance_port      = 80
    instance_protocol  = "http"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = "${data.aws_acm_certificate.neworca.arn}"
  }

  listener = {
    instance_port     = 22
    instance_protocol = "tcp"
    lb_port           = 22
    lb_protocol       = "tcp"
  }

  health_check = {
    target              = "HTTP:80/"
    timeout             = 5
    interval            = 30
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }

  instances = ["${aws_instance.svc-instance.id}"]

  tags = {
    Name = "${var.vpc_name}"
  }
}

################################
# EC2
################################

resource "aws_instance" "svc-instance" {
  ami           = "${data.aws_ami.amazon_linux.id}"
  instance_type = "t2.micro"

  key_name  = "${var.ssh_key_name}"
  subnet_id = "${aws_subnet.private.id}"

  security_groups = [
    "${aws_security_group.vpc-sg.id}",
    "${aws_security_group.svc-sg.id}",
  ]

  tags = {
    Name = "${var.vpc_name}-svc-instance"
  }
}

resource "null_resource" "provision-and-run" {
  depends_on = ["aws_instance.svc-instance"]

  triggers = {
    svc_instance_ids = "${join(",", aws_instance.svc-instance.*.id)}"
  }

  provisioner "remote-exec" {
    connection {
      type        = "ssh"
      user        = "ec2-user"
      host        = "${aws_elb.elb.dns_name}"
      private_key = "${var.ssh_key_contents}"
    }

    inline = [
      "sudo yum update -y",
      "sudo yum install golang",
      "git clone https://github.com/tlake/tfe-test.git",
      "cd tfe-test/src/",
      "go build .",
      "nohup ./src &",
    ]
  }
}
