SERVICE="blobfuse"
SCRIPT="stressTest"

# crontab -e and at the end enter below line to run it on every 00 and 30 minutes of every hour
# m h  dom mon dow   command
#0,30 * * * * /home/vibhansa/stress_test.sh

if pgrep -x "$SERVICE" > /dev/null
then
        if pgrep -x "$SCRIPT" > /dev/null
        then
                echo "`date` :: Already running" >> /home/vibhansa/stress_test.log
        else
                if [ `stat -c %s /home/vibhansa/stress_test.log` -gt 10485760 ]
                then
                        echo "`date` :: Trimmed " > /home/vibhansa/stress_test.log
                fi

                echo "`whoami` : `date` :: Starting stress test " >> /home/vibhansa/stress_test.log

                mem=$(top -b -n 1 -p `pgrep -x blobfuse` | tail -1)
                echo $mem >> /home/vibhansa/stress_test.log

                rm -rf /home/vibhansa/blob_mnt/stress
                go run /home/vibhansa/code/azure-storage-fuse/test/stressTest.go /home/vibhansa/blob_mnt
                echo "`whoami` : `date` :: Ending stress test " >> /home/vibhansa/stress_test.log
                cp  /home/vibhansa/stress_test.log  /home/vibhansa/blob_mnt/
        fi
else
        echo "`date` :: Re-Starting blobfuse" >> /home/vibhansa/stress_test.log
        rm -rf /home/vibhansa/blob_mnt/*
        sudo ~/clear.sh
        sudo fusermount -u ~/blob_mnt
        /home/vibhansa/mnt.sh

        if [ `stat -c %s /home/vibhansa/blob_mnt/restart` -gt 10485760 ]
        then
                echo "`date` Trimmed " > /home/vibhansa/blob_mnt/restart
        fi
        echo "`date` Restart" >> /home/vibhansa/blob_mnt/restart

        cp /var/log/blobfuse.log /home/vibhansa/blob_mnt/
fi

