# MEMBER SERVICES
resource "aws_ecs_task_definition" "deployer" {
  family = "deployer"
  container_definitions = "${file("tasks/deployer.json")}"

  volume {
    name = "chaincerts"
    host_path = "/home/ec2-user/keys"
  }
}

resource "aws_ecs_service" "deployer" {
  name = "deployer"
  cluster = "${aws_ecs_cluster.netvote.id}"
  task_definition = "${aws_ecs_task_definition.deployer.arn}"
  desired_count = 1
  depends_on = ["aws_iam_role.ecs_instance_role", "aws_ecs_service.rootpeer"]
}