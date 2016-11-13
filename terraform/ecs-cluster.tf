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