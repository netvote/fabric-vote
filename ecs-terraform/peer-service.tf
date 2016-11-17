variable "peer_count" {
  default = 3
}

# validating peer SERVICES

resource "aws_ecs_task_definition" "peer" {
  family = "peer${count.index}"
  container_definitions = "${replace(file("tasks/vpeer.json"), "_NODEINDEX_", count.index+1)}"
  count = "${var.peer_count}"
  volume {
    name = "dockersock"
    host_path = "/var/run/docker.sock"
  }
}

resource "aws_ecs_service" "peer" {
  name = "vpeer${count.index}"
  cluster = "${aws_ecs_cluster.netvote.id}"
  task_definition = "${element(aws_ecs_task_definition.peer.*.arn, count.index)}"
  desired_count = 1
  iam_role = "${aws_iam_role.ecs_instance_role.arn}"
  count = "${var.peer_count}"
  depends_on = ["aws_iam_role.ecs_instance_role"]
  load_balancer {
    elb_name = "${element(aws_elb.peer.*.name, count.index)}"
    container_name = "vpeer${count.index+1}"
    container_port = 7050
  }
}


resource "aws_route53_record" "peer" {
  zone_id = "ZTO1AJYOEZG73"
  name = "vp${count.index+1}.peer"
  count = "${var.peer_count}"
  type = "CNAME"
  records = ["${element(aws_elb.peer.*.dns_name, count.index)}"]
  ttl = "30"
}

resource "aws_elb" "peer" {
  name = "peer-service-elb-${count.index}"
  availability_zones = ["us-east-1a","us-east-1c","us-east-1d"]
  count = "${var.peer_count}"

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