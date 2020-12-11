#!/bin/bash

mnt_path="/home/vibhansa/blob_mnt"
#mnt_path="/testhdd/blob_mnt"

echo "Using : " `~/blob --version`
echo "Mounting blobfuse to : $mnt_path"

if [ "$1" == "adls" ]
then
    echo "ADLS MODE ON....."
	#~/blob $mnt_path --tmp-path=/mnt/blobfusetmp --config-file=/home/vibhansa/myblob.cfg.vikasfusename -o allow_other --use-adls=true --use-attr-cache=true --log-level=LOG_DEBUG
	~/blob $mnt_path --tmp-path=/mnt/ramdisk --config-file=/home/vibhansa/myblob.cfg.vikasfusename -o allow_other --use-adls=true --use-attr-cache=true --log-level=LOG_DEBUG
elif [ "$1" == "highmem" ]
then
    echo "HIGHMEM MODE ON....."
	sudo ~/blob $mnt_path --tmp-path=/mnt/ramdisk --config-file=/home/vibhansa/myblob.cfg -o allow_other --file-cache-timeout-in-seconds=120 --use-attr-cache=true -o attr_timeout=1 -o entry_timeout=1 -o negative_timeout=120
elif [ "$1" == "noc" ]
then
    echo "Block MODE NOCACHE ON....."
	~/blob $mnt_path --tmp-path=/mnt/blobfusetmp --config-file=/home/vibhansa/myblob.cfg -o allow_other --file-cache-timeout-in-seconds=0 --log-level=LOG_DEBUG
elif [ "$1" == "124" ]
then
    echo "Block MODE ON 124....."
	/home/vibhansa/blobfuse/vibhansa-msft/v1.2.4/azure-storage-fuse/build/blobfuse $mnt_path --tmp-path=/mnt/blobfusetmp --config-file=/home/vibhansa/myblob.cfg -o allow_other --file-cache-timeout-in-seconds=0 --use-attr-cache=true --log-level=LOG_DEBUG
elif [ "$1" == "attr" ]
then
    echo "Block MODE ON Attr cache invalidate....."
	~/blob $mnt_path --tmp-path=/mnt/blobfusetmp --config-file=/home/vibhansa/myblob.cfg -o allow_other --file-cache-timeout-in-seconds=120 --attr-cache-timeout-in-seconds=120 --use-attr-cache=true --log-level=LOG_DEBUG
elif [ "$1" == "old" ]
then
    echo "Block MODE ON OLDER....."
	~/blob $mnt_path --tmp-path=/mnt/blobfusetmp --config-file=/home/vibhansa/myblob.cfg -o allow_other --file-cache-timeout-in-seconds=120 -o attr_timeout=1 -o entry_timeout=1 -o negative_timeout=120 --log-level=LOG_DEBUG
else
    echo "Block MODE ON....."
	~/blob $mnt_path --tmp-path=/mnt/blobfusetmp --config-file=/home/vibhansa/myblob.cfg -o allow_other --file-cache-timeout-in-seconds=120 --use-attr-cache=true -o attr_timeout=1 -o entry_timeout=1 -o negative_timeout=120 --log-level=LOG_DEBUG
fi
