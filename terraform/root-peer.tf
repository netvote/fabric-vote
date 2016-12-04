resource "aws_instance" "root-peer" {
  ami = "${var.docker_ami}"
  instance_type = "${var.instance_size}"
  key_name = "netvote-slanders"

  tags {
    Name = "root-peer"
    System = "netvote-hyperledger"
  }

  user_data = "${replace(file("conf/userdata-ecs-cluster.txt"), "CLUSTER_NAME", aws_ecs_cluster.netvote.name)}"

  iam_instance_profile = "${aws_iam_instance_profile.ecs_instance_profile.id}"
  depends_on = ["aws_instance.membership"]

  provisioner "file" {
    source = "root-peer"
    destination = "/home/ec2-user"
    connection {
      type = "ssh"
      user = "ec2-user"
      private_key = "${file(var.keyfile)}"
    }
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /home/ec2-user/root-peer/init.sh",
      "sudo /home/ec2-user/root-peer/init.sh"
    ]
    connection {
      type = "ssh"
      user = "ec2-user"
      private_key = "${file(var.keyfile)}"
    }
  }
}

resource "aws_elb" "root" {
  name = "root-peer-elb"
  availability_zones = ["us-east-1a","us-east-1c","us-east-1d"]

  instances = ["${aws_instance.root-peer.id}"]

  listener {
    instance_port = 7050
    instance_protocol = "tcp"
    lb_port = 80
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

resource "aws_route53_record" "root-peer" {
  zone_id = "${var.route_53_zone_id}"
  name = "root"
  type = "CNAME"
  records = ["${aws_instance.root-peer.private_dns}"]
  ttl = "30"
}

resource "aws_route53_record" "root-peer-pub" {
  zone_id = "${var.route_53_zone_id}"
  name = "peer"
  type = "CNAME"
  records = ["${aws_elb.root.dns_name}"]
  ttl = "30"
}

# ECS RESOURCES
resource "aws_ecs_cluster" "netvote" {
  name = "netvote-fabric"
}

resource "aws_ecs_task_definition" "deployer" {
  family = "deployer"
  container_definitions = "${file("tasks/deployer.json")}"
  depends_on = ["aws_instance.peer"]
  volume {
    name = "chaincerts"
    host_path = "/var/deployer/keyvalstore"
  }
}

resource "aws_ecs_service" "deployer" {
  name = "deployer"
  cluster = "${aws_ecs_cluster.netvote.id}"
  task_definition = "${aws_ecs_task_definition.deployer.arn}"
  desired_count = 1
  depends_on = ["aws_instance.peer"]
}


resource "aws_ecs_task_definition" "chainapi" {
  family = "chain-api"
  container_definitions = "${file("tasks/chainapi.json")}"

  volume {
    name = "chaincerts"
    host_path = "/var/chain_api/keyvalstore"
  }
}


resource "aws_ecs_service" "chainapi" {
  name = "chain-api"
  cluster = "${aws_ecs_cluster.netvote.id}"
  task_definition = "${aws_ecs_task_definition.chainapi.arn}"
  desired_count = 1
  iam_role = "${aws_iam_role.ecs_instance_role.arn}"
  depends_on = ["aws_instance.peer"]
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