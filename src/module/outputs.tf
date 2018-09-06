output "deploy_id" {
  value = "${layer0_deploy.guestbook.id}"
}

output "load_balancer_id" {
  value = "${layer0_load_balancer.guestbook.id}"
}

output "load_balancer_url" {
  value = "http://${layer0_load_balancer.guestbook.url}"
}

output "service_id" {
  value = "${layer0_service.guestbook.id}"
}
