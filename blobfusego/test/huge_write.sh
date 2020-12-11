
#!/bin/bash


mkdir ~/blob_mnt/$1

#for i in {1..800000}
#do
#	echo "Writeing " $i.tst
#	time dd if=/dev/zero of=~/blob_mnt/$1/abcdefghijklmnopqrstuvwxyz_1234567890_ABCDEFGHIJKLMNOPQRSTUVWXYZ_$i.tst bs=1M count=1
#done

#for i in {1..200000}
#do
#	echo "Writeing " $i.tst
#	time dd if=/dev/zero of=~/blob_mnt/$1/abcdefghijklmnopqrstuvwxyz_1234567890_$i.tst bs=1M count=1
#done

for i in {1..5000}

do
	echo "Writeing " $i.tst
	time dd if=/dev/zero of=~/blob_mnt/$1/abcdefghijklmnopqrstuvwxyz_$i.tst bs=1M count=1
done
