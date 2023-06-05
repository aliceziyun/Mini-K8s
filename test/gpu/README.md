

### 提交GPU程序并获取结果
从本地/home/data/gpu/目录复制matrix.cu和matrix.slurm(需要事先创好)到remote的/Job/gpu目录（即 `WorkDir` ），
然后到 `WorkDir` 下Sbtach，生成的.err和.out文件就在 `WorkDir` ；
server会定期（10s一次）调用 `sacct` 命令查看刚才提交的job的状态，如果是COMPLETED，则将工作目录下的.err和.out文件复制到本地的/home/data/res目录（会在res目录下创一个 `"res"+jobId` 目录，这个job的.err和.out文件会放在里面）

还得考虑镜像，以及如何将结果返回到镜像外。