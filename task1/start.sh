# Starting $1 docker containers

(( port = 8081 ))

for (( i = 0; i < $1; i++ ))
do
	docker run -d -p $port:22 sshd
	(( port += 1 ))
done

