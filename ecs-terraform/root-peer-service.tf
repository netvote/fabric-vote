# rootpeer SERVICES


resource "aws_ecs_task_definition" "rootpeer" {
  family = "rootpeer"
  container_definitions = "${file("tasks/rootpeer.json")}"

  volume {
    name = "dockersock"
    host_path = "/var/run/docker.sock"
  }
}

resource "aws_ecs_service" "rootpeer" {
  name = "rootpeer"
  cluster = "${aws_ecs_cluster.netvote.id}"
  task_definition = "${aws_ecs_task_definition.rootpeer.arn}"
  desired_count = 1
  iam_role = "${aws_iam_role.ecs_instance_role.arn}"
  depends_on = ["aws_iam_role.ecs_instance_role","aws_ecs_service.memberservice"]
  load_balancer {
    elb_name = "${aws_elb.rootpeer.name}"
    container_name = "rootpeer"
    container_port = 7050
  }
}

resource "aws_route53_record" "rootpeer" {
  zone_id = "ZTO1AJYOEZG73"
  name = "peer"
  type = "CNAME"
  records = ["${aws_elb.rootpeer.dns_name}"]
  ttl = "30"
}

resource "aws_elb" "rootpeer" {
  name = "rootpeer-service-elb"
  availability_zones = ["us-east-1a","us-east-1c","us-east-1d"]

  listener {
    instance_port = 7050
    instance_protocol = "tcp"
    lb_port = 7050
    lb_protocol = "tcp"
  }

  listener {
    instance_port = 9051
    instance_protocol = "tcp"
    lb_port = 7051
    lb_protocol = "tcp"
  }

  listener {
    instance_port = 9052
    instance_protocol = "tcp"
    lb_port = 7052
    lb_protocol = "tcp"
  }

  listener {
    instance_port = 9375
    instance_protocol = "tcp"
    lb_port = 2375
    lb_protocol = "tcp"
  }

  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 3
    timeout = 20
    target = "TCP:7050"
    interval = 30
  }
}