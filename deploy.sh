
for CC in ./*/;
do
  CC=$(basename $CC)
  tar -czf ./${CC}/code.tar.gz ./${CC}/connection.json
  tar -czf ./cealgull_${CC}.tar.gz ./${CC}/code.tar.gz ./${CC}/metadata.json
done
