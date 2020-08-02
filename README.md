# Go 静态文件 Web 服务

因为使用 [MinIO](https://github.com/minio/minio) 作为图片管理服务，  
但 [MinIO](https://github.com/minio/minio) 没有提供方便的文件浏览服务，  
所以编写了本引用~

## 功能
除了能浏览文件外，如果是图片文件  
假设你有一张图片地址是： http://127.0.0.1:8080/img/name.png  
那么你可以浏览： http://127.0.0.1:8080/img/name.png_180x180.png  
从而生成一张缩略图，来减少网络传输~