# basic script to aid ssh

node_ip=`ecs-terraform show |grep public_ip | sed -e "${1}q;d" |sed 's/.*= //g'`

ssh -oStrictHostKeyChecking=no ec2-user@$node_ip -i ~/.ssh/netvote-slanders.pem
