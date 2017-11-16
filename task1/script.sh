# Integrate function x^2 from $1 to $2
# $1 -- start position
# $2 -- end position
# $3 -- count steps

pos=$(bc <<< "$1")
step_size=$(bc <<< "scale=10; ($2 - $1) / $3")
(( square = 0 ))
for (( step=0; step < $3; step++))
do
	square=$(bc <<< "scale=10; $square+0.5*($pos*$pos+($pos+$step_size)*($pos+$step_size))*$step_size")
	#echo "print $pos $square"
	pos=$(bc <<< "scale=10; $pos+$step_size")
done
echo $square

