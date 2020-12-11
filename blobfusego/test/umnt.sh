#!/bin/bash

mnt_path="/home/vibhansa/blob_mnt"
#mnt_path="/testhdd/blob_mnt"

if [ $# -eq 1 ]
then
    mnt_path=$1
fi

echo "Un-Mounting blobfuse from : $mnt_path"

sudo fusermount -u $mnt_path
