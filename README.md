make&& time ./bin/knload --stages=10:30  --service-url=httpload-server.default.example.com?sleep=200 --gateway-address=39.97.31.219:80 --save-path=tmp/test.html --namespace=default --label='app=httpload-server' -v 5

time /app/bin/knload --stages=3:10,10:30,50:30,100:30,450:50  --service-url=httpload-server.default.example.com?sleep=400 --gateway-address=39.97.31.219:80 --save-path=/tmp/index.html --namespace=default --label='app=httpload-server' -v 5

