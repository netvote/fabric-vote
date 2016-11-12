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

# ECS RESOURCES
resource "aws_instance" "ecs-cluster" {
  ami = "ami-46134b51"
  instance_type = "m3.large"
  count = 1
  iam_instance_profile = "${aws_iam_instance_profile.ecs_instance_profile.id}"
  key_name = "netvote-slanders"
  user_data = "${replace(file("conf/userdata-ecs-cluster.txt"), "CLUSTER_NAME", aws_ecs_cluster.netvote.name)}"
}

# ECS RESOURCES
resource "aws_ecs_cluster" "netvote" {
  name = "netvote-fabric"
}

resource "aws_iam_role" "ecs_instance_role" {
  name = "ecsInstanceRole"
  assume_role_policy = "${file("conf/ecs_instance_role_trust.json")}"
}

resource "aws_iam_policy_attachment" "ecs_service_role" {
  name = "ecs_service_role"
  roles = ["${aws_iam_role.ecs_instance_role.name}"]
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceRole"
}

resource "aws_iam_policy_attachment" "ecs_service_ec2_role" {
  name = "ecs_service_ec2_role"
  roles = ["${aws_iam_role.ecs_instance_role.name}"]
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_profile" {
  name = "ecsInstanceRole"
  roles = ["${aws_iam_role.ecs_instance_role.name}"]
}