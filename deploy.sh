# scp -r public/* jwebsite:/var/www/html/ffreplay
# use rsync to copy files, ignore file already in the server
rsync -avz public/* jwebsite:/var/www/html/ffreplay