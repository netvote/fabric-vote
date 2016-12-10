variable "keyfile" {
  type = "string",
  description = "pem file for building EC2 Instances"
}

variable "access_key" {
  type = "string",
}

variable "secret_key" {
  type = "string",
}

variable "docker_ami" {
  default="ami-46134b51"
}

variable "instance_size" {
  default="m3.large"
}

variable "route_53_zone_id" {
  default="ZTO1AJYOEZG73"
}

variable "region" {
  type = "string"
}

variable "account"{
  type = "string"
}

provider "aws" {
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region = "us-east-1"
}

resource "aws_iam_role" "ecs_instance_role" {
  name = "ecsInstanceRole"
  assume_role_policy = "${file("conf/ec2_instance_role_trust.json")}"
  provisioner "local-exec" {
    command = "sleep 60"
  }
}

resource "aws_iam_policy_attachment" "ecs_service_role" {
  name = "ecs_service_role"
  roles = ["${aws_iam_role.ecs_instance_role.name}"]
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceRole"
  provisioner "local-exec" {
    command = "sleep 60"
  }
}

resource "aws_iam_policy_attachment" "ecs_service_ec2_role" {
  name = "ecs_service_ec2_role"
  roles = ["${aws_iam_role.ecs_instance_role.name}"]
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
  provisioner "local-exec" {
    command = "sleep 60"
  }
}

resource "aws_iam_instance_profile" "ecs_instance_profile" {
  name = "ecsInstanceRole"
  roles = ["${aws_iam_role.ecs_instance_role.name}"]
  provisioner "local-exec" {
    command = "sleep 60"
  }
}