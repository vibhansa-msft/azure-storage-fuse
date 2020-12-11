~/umnt.sh
sudo ~/clear.sh
rm -rf blobfuse.log

echo "Cleanup done now mounting"
if [ "$1" == "sas" ]
then
    echo "SAS MODE ON....."
	go run blobfuse.go -mount-path="/home/vibhansa/blob_mnt" -tmp-path="/mnt/ramdisk" -fs=azurestorage -fd=gofuse -log-level=LOG_ERR -log-file=/home/vibhansa/blobfuse.log -account=vikasfuseblob -sas="?sv=2019-12-12&ss=b&srt=sco&sp=rwlacx&se=2021-09-29T14:43:37Z&st=2020-09-29T06:43:37Z&spr=https,http&sig=Mr1TUk3m%2B6l0YmphFsJ6%2BROFr%2BrNzoypsti1gFWsXzk%3D" -authtype=sas -container=testcntgo -block-size-in-mb=8 -parallelism=64 
elif [ "$1" == "ram" ]
then
    echo "RAMDISK MODE ON....."
	go run blobfuse.go -mount-path="/home/vibhansa/blob_mnt" -tmp-path="/mnt/ramdisk" -fs=azurestorage -fd=gofuse -log-level=LOG_ERR -log-file=/home/vibhansa/blobfuse.log -account=vikasfuseblob -accountkey=B6EQf3MbdIN1VGtYCrY9vs8pTLrNGCniRX+GAMx8t7RE00ZJFlFNsy+/nbq9sbwDUIxNnZbiOsb4/EcDnwDasQ== -authtype=key -container=testcntgo -block-size-in-mb=8 -parallelism=128 
elif [ "$1" == "gofuse" ]
then
    echo "GOFUSE MODE ON....."
	go run blobfuse.go -mount-path="/home/vibhansa/blob_mnt" -tmp-path="/mnt/blobfusetmp" -fs=azurestorage -fd=gofuse -log-level=LOG_DEBUG -log-file=/home/vibhansa/blobfuse.log -account=vikasfuseblob -accountkey=B6EQf3MbdIN1VGtYCrY9vs8pTLrNGCniRX+GAMx8t7RE00ZJFlFNsy+/nbq9sbwDUIxNnZbiOsb4/EcDnwDasQ== -authtype=key -container=testcntgo 
else
    echo "NORMAL MODE ON....."
	go run blobfuse.go -mount-path="/home/vibhansa/blob_mnt" -tmp-path="/mnt/ramdisk" -fs=azurestorage -fd=bazil -log-level=LOG_DEBUG -log-file=/home/vibhansa/blobfuse.log -account=vikasfuseblob -accountkey=B6EQf3MbdIN1VGtYCrY9vs8pTLrNGCniRX+GAMx8t7RE00ZJFlFNsy+/nbq9sbwDUIxNnZbiOsb4/EcDnwDasQ== -authtype=key -container=testcntgo 
fi

