
#!/bin/bash

rm -rf /home/vibhansa/blob_mnt/testfile*

#time dd if=/dev/zero of=/home/vibhansa/blob_mnt/testfile200 bs=1M count=200
#time dd if=/dev/zero of=/home/vibhansa/blob_mnt/testfile500 bs=1M count=500
time dd if=/dev/zero of=/home/vibhansa/blob_mnt/testfile0_1 bs=1M count=1024 oflag=direct
time dd if=/dev/zero of=/home/vibhansa/blob_mnt/testfile0_2 bs=1M count=1024 oflag=direct
time dd if=/dev/zero of=/home/vibhansa/blob_mnt/testfile0_3 bs=1M count=1024 oflag=direct

for i in {1,2,3,4,5,6,7,8,9,10}
do
    echo "-----------------------------------------------------"
    ~/write.sh $i
    rm -rf /mnt/blobfusetmp/root/*
    ~/read.sh $i
    rm -rf /home/vibhansa/blob_mnt/test*
done
