resource "aws_instance" "membership" {
  ami = "${var.docker_ami}"
  instance_type = "${var.instance_size}"
  key_name = "netvote-slanders"

  tags {
    Name = "membership-service"
    System = "netvote-hyperledger"
  }

  provisioner "file" {
    source = "membership"
    destination = "/home/ec2-user"
    connection {
      type = "ssh"
      user = "ec2-user"
      private_key = "${file(var.keyfile)}"
    }
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /home/ec2-user/membership/init.sh",
      "mkdir -p /var/hyperledger/production/db",
      "sudo /home/ec2-user/membership/init.sh"
    ]
    connection {
      type = "ssh"
      user = "ec2-user"
      private_key = "${file(var.keyfile)}"
    }
  }
}

resource "aws_route53_record" "members" {
  zone_id = "${var.route_53_zone_id}"
  name = "members"
  type = "CNAME"
  records = ["${aws_instance.membership.private_dns}"]
  ttl = "30"
}