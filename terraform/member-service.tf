# MEMBER SERVICES
resource "aws_ecs_task_definition" "memberservice" {
  family = "member-service"
  container_definitions = "${file("tasks/membersrvc.json")}"
}

resource "aws_ecs_service" "memberservice" {
  name = "member-service"
  cluster = "${aws_ecs_cluster.netvote.id}"
  task_definition = "${aws_ecs_task_definition.memberservice.arn}"
  desired_count = 1
  iam_role = "${aws_iam_role.ecs_instance_role.arn}"

  load_balancer {
    elb_name = "${aws_elb.memberservice.name}"
    container_name = "member-service"
    container_port = 7054
  }
}

resource "aws_route53_record" "memberservice" {
  zone_id = "ZTO1AJYOEZG73"
  name = "members"
  type = "CNAME"
  records = ["${aws_elb.memberservice.dns_name}"]
  ttl = "30"
}

resource "aws_elb" "memberservice" {
  name = "membership-service-elb"
  availability_zones = ["us-east-1a","us-east-1c","us-east-1d"]

  listener {
    instance_port = 7054
    instance_protocol = "tcp"
    lb_port = 7054
    lb_protocol = "tcp"
  }

  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 10
    timeout = 20
    target = "TCP:7054"
    interval = 30
  }
}