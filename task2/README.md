# Задние 2


## Постановка задачи

Реализовать алгоритм Gossip и измерить скорость распространения сообщения по сети в 
зависимости от доли потерь udp пакетов в сети

## Имитация сброса пакетов

1. Для сброса пакетов используем команду

		iptables -A INPUT -p udp -i lo -m statistic --mode random --probability 0.1 -j DROP

1. Для восстановления пропускной способности используем команду
		
		iptables -D INPUT -p udp -i lo -m statistic --mode random --probability 0.1 -j DROP
