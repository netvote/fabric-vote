# MEMBER SERVICES
resource "aws_ecs_task_definition" "chainapi" {
  family = "chain-api"
  container_definitions = "${file("tasks/chainapi.json")}"

  volume {
    name = "chaincerts"
    host_path = "/home/ec2-user/keys"
  }
}

resource "aws_ecs_service" "chainapi" {
  name = "chain-api"
  cluster = "${aws_ecs_cluster.netvote.id}"
  task_definition = "${aws_ecs_task_definition.chainapi.arn}"
  desired_count = 1
  iam_role = "${aws_iam_role.ecs_instance_role.arn}"
  depends_on = ["aws_iam_role.ecs_instance_role", "aws_ecs_service.rootpeer"]
  load_balancer {
    elb_name = "${aws_elb.chainapi.name}"
    container_name = "chain-api"
    container_port = 8000
  }
}

resource "aws_route53_record" "chainapi" {
  zone_id = "ZTO1AJYOEZG73"
  name = "accounts"
  type = "CNAME"
  records = ["${aws_elb.chainapi.dns_name}"]
  ttl = "30"
}

resource "aws_elb" "chainapi" {
  name = "chainapi-service-elb"
  availability_zones = ["us-east-1a","us-east-1c","us-east-1d"]

  listener {
    instance_port = 8000
    instance_protocol = "tcp"
    lb_port = 80
    lb_protocol = "tcp"
  }

  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 3
    timeout = 20
    target = "TCP:8000"
    interval = 30
  }
}