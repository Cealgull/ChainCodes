
for CC in ./*/;
do
  CC=$(basename $CC)
  cd ${CC}
  tar -czf code.tar.gz connection.json
  tar -czf cealgull_${CC}.tar.gz code.tar.gz metadata.json
  cd ..
  mv ${CC}/cealgull_${CC}.tar.gz .
done
