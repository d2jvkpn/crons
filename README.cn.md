### Crons
---
Crons 程序用于控制一组命令行的定时启动或重启, 支持 windows, linux 平台。

#### 1. Crons
##### 1.1 安装与编译
```bash
git clone git@github.com:d2jvkpn/crons.git
cd crons

make build
ls target/crons*
```

##### 1.2 配置文件 (yaml)
```yaml
jobs:
# required: name, path, cron
- name: SLEEP
  path: sleep
  args: [70]
  working_dir: ""
  # https://crontab.guru/#00_00_*_*_1
  cron:
    minute: "*"
    hour: "*"
    month_day: "*"
    month: "*"
    week_day: "*/3"
  max_retries: 2
  start_immediately: true
```

jobs 下配置一组任务
- *name: string, 任务名称;
- *path: string, 执行文件名称或路径;
- args: []string([]string{}), 命令参数列表;
- working_dir: string(""), 工作目录;
- cron: 定时配置 (https://crontab.guru/#00_00_*_*_1)
  - minute: string("*"), 分钟;
  - hour: string("*"), 小时;
  - month_day: string("*"), 日期;
  - month: string("*"), 月份;
  - week_day: string("*"), 星期 (0 表示周日);
- max_retries: uint(0), 运行失败重试次数;
- start_immediately: bool(false), 是否直接启动 (即不等待下一次 cron 时间);

##### 1.3 运行
```bash
./target/crons -config configs/local.yaml
```

#### 2. web
待开发...
