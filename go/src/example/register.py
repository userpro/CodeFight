import time
import random
from CodeWar.Utils import CodeWar

# 账号及服务器设置 (C)
chrome = '' # 对于Windows用户可能需要填写Chrome安装路径.../chrome.exe
username = 'test3'
password = 'test3'
email = 'test'
url = '127.0.0.1'
port = '52333'

if __name__ == '__main__':
    # 此处不需要修改
    my = CodeWar(url=url, port=port, 
        username=username, password=password, email=email, 
        chrome=chrome)

    my.register()
