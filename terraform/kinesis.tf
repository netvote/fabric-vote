resource "aws_kinesis_stream" "votes" {
  name = "votes"
  shard_count = 2
  retention_period = 48
  shard_level_metrics = [
    "IncomingBytes",
    "OutgoingBytes"
  ]
}