# gugo
简易的静态博客网站生成器

## 使用方法
1. 下载我的项目
```
git clone github.com/qwendy/gugo
```

2. 下载glide，用于安装所有依赖的包

```
go get github.com/Masterminds/glide
```
安装好glide之后,安装需要的依赖
```
glide install
```

3. 运行
在项目的主文件夹创建source/_post文件夹。将你的markdown文件放在此文件夹内。运行
```
go run main.go
```
使用
```
bash push.sh
```
上传静态博客文件夹至github