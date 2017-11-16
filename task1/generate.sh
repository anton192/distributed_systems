# $1 - count of ranges
# $2 - all intervals

(( pos=0 ))
programm_size=$(bc <<< "scale=10; 1/$1")
count_steps_in_programm=$(bc <<< "scale=0; $2/$1")
for (( i=0; i<$1; i++ ))
do
	new_pos=$(bc <<< "scale=10; $pos+$programm_size")
	echo "$pos $new_pos $count_steps_in_programm"
	pos=$new_pos
done

