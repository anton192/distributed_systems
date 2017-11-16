
# Init $1 ssh docker containers (with names like sshd<i>)

cp ./.ssh/test_rsa ~/.ssh/docker_sshd
echo '' | cat > my_cluster

for (( i=0; i<$1; i++ ))
do
	(( port = $i + 8081 ))
	echo "Host sshd-$i" >> ~/.ssh/config
	echo "	User root" >> ~/.ssh/config
	echo "	HostName localhost" >> ~/.ssh/config
	echo "	Port $port" >> ~/.ssh/config
	echo "	IdentityFile ~/.ssh/docker_sshd" >> ~/.ssh/config

	echo "sshd-$i" >> my_cluster
done

cp ~/.ssh/config ~/.ssh/config.backup
(echo 'Host *'; echo StrictHostKeyChecking no) >> ~/.ssh/config
parallel --slf my_cluster --nonall true
mv ~/.ssh/config.backup ~/.ssh/config

rm my_cluster
