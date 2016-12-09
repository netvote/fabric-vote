//resource "aws_instance" "peer" {
//  ami = "${var.docker_ami}"
//  instance_type = "${var.instance_size}"
//  key_name = "netvote-slanders"
//
//  count = 3
//
//  tags {
//    Name = "peer"
//    System = "netvote-hyperledger"
//  }
//
//  user_data = "${replace(file("conf/userdata-ecs-cluster.txt"), "CLUSTER_NAME", "default")}"
//
//  depends_on = ["aws_instance.membership", "aws_instance.root-peer"]
//
//  provisioner "file" {
//    source = "peer"
//    destination = "/home/ec2-user"
//    connection {
//      type = "ssh"
//      user = "ec2-user"
//      private_key = "${file(var.keyfile)}"
//    }
//  }
//
//  provisioner "remote-exec" {
//    inline = [
//      "chmod +x /home/ec2-user/peer/init.sh",
//      "mkdir -p /var/hyperledger/production/db",
//      "sudo /home/ec2-user/peer/init.sh vp${count.index + 2}"
//    ]
//    connection {
//      type = "ssh"
//      user = "ec2-user"
//      private_key = "${file(var.keyfile)}"
//    }
//  }
//}