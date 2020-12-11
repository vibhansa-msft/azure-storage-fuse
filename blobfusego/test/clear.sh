kill -9 `pidof blob`
fusermount -u ~/blob_mnt
sudo umount -f ~/blob_mnt

rm -rf /mnt/blobfusetmp/*
rm -rf ~/blob_mnt/*
rm -rf /mnt/ramdisk/*
