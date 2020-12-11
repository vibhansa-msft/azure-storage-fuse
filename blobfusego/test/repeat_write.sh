
#!/bin/bash


out_file="/home/vibhansa/blob_mnt/a.txt"

while [ 1 ] 
do
	echo `date` >> $out_file
	echo `date`
	sleep 10
done
