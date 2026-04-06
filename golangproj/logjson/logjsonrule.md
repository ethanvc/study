# 基本规则
1. internal/gostd 目录下以logjson开头的文件，是本模块个性的功能，其余为从标准库的json部分同步过来的，同步的函数为copy_jsonv2.py。
2. 修改代码时，尽可能改动以logjson开头的文件。