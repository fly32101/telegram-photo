@echo off
echo 正在构建前端应用...
cd frontend
call npm run build

echo 创建后端静态文件目录...
cd ..
mkdir backend\dist 2>nul

echo 复制前端构建产物到后端静态目录...
xcopy /E /Y frontend\dist\* backend\dist\

echo 部署完成！
echo 您可以使用 main_with_frontend.go 替换 main.go 来启动集成了前端的后端服务
echo 或者直接运行: cd backend && go run main_with_frontend.go