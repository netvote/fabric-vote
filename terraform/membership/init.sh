curl -L "https://github.com/docker/compose/releases/download/1.8.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

mkdir -p /var/hyperledger/production/db

cd /home/ec2-user/membership/

echo "awaiting docker...sleep 30..."
sleep 30

sed -i "s/IP_ADDRESS/$(curl http://169.254.169.254/latest/meta-data/local-ipv4)/g" docker-compose.yml

/usr/local/bin/docker-compose up -d