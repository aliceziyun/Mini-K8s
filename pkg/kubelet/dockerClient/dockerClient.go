package dockerClient

import (
	"Mini-K8s/pkg/kubelet/message"
	"Mini-K8s/pkg/object"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"unsafe"

	"github.com/docker/docker/api/types/mount"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const ServiceDns string = "10.10.10.10"

// 获取一个新Client。可以被外部引用。
func GetNewClient() (*client.Client, error) {
	return client.NewClientWithOpts()
}

// 根据containerId 获取其信息
func getContainerInfo(containerId string) (containerInfo types.ContainerJSON, err error) {
	cli, err2 := GetNewClient()
	if err2 != nil {
		fmt.Println(err2)
	}
	containerJson, err := cli.ContainerInspect(context.Background(), containerId)
	if err != nil {
		fmt.Println("error")
		fmt.Println(err)
	}
	fmt.Printf(
		"=======容器信息======\nID:%+v\nname:%+v\nimage:%+v\n",
		// containerJson.ID[:10],
		containerJson.ID,
		containerJson.Name,
		containerJson.Image,
	)
	return containerJson, err
}

// 获取所有Containers
func getAllContainers() ([]types.Container, error) {
	fmt.Println("[getAllContainers]:")
	cli, err2 := GetNewClient()
	if err2 != nil {
		fmt.Println(err2)
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		fmt.Println(err)
	}
	for _, container := range containers {
		fmt.Printf("id=%s, name=%s, image=%s\n", container.ID[:10], container.Names, container.Image)
		fmt.Printf("status=%s (?)\n", container.Status)
	}
	return containers, err
}

// 根据指定containerId删除指定容器（单个）
func deleteContainerById(containerId string) error {
	cli, err := GetNewClient()
	if err != nil {
		return err
	}
	err = cli.ContainerStop(context.Background(), containerId, nil)
	if err != nil {
		fmt.Println("error on Stopping a Container")
		return err
	}
	err = cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{})
	if err != nil {
		fmt.Println("error on Removing a Container")
		return err
	}
	return nil
}

// 根据指定containerIds删除指定容器（多个）
func deleteContainersByIds(containerIds []string) error {
	cli, err2 := GetNewClient()
	if err2 != nil {
		return err2
	}
	//需要先停止containers
	for _, value := range containerIds {
		err := cli.ContainerStop(context.Background(), value, nil)
		if err != nil {
			fmt.Println("error on Stopping a Container")
			return err
		}
	}
	//停止后删除
	for _, value := range containerIds {
		err := cli.ContainerRemove(context.Background(), value, types.ContainerRemoveOptions{})
		if err != nil {
			fmt.Println("error on Removing a Container")
			return err
		}
	}
	return nil
}

// 获取运行中的容器
func getRunningContainers() ([]types.Container, error) {
	cli, err := GetNewClient()
	if err != nil {
		fmt.Println(err)
	}
	return cli.ContainerList(context.Background(), types.ContainerListOptions{})
}

// 启动指定id的容器
func startContainer(containerId string) error {
	cli, err := GetNewClient()
	//cli, err := client.NewClientWithOpts()
	if err != nil {
		// fmt.Println(err)
		return err
	}
	err = cli.ContainerStart(context.Background(), containerId, types.ContainerStartOptions{})
	return err
}

// 停止指定id的容器
func stopContainer(containerId string) error {
	cli, err := GetNewClient()
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = cli.ContainerStop(context.Background(), containerId, nil)
	return err
}

// 创建pause容器
func createPause(ports []object.ContainerPort, name string) (container.ContainerCreateCreatedBody, error) {
	fmt.Println("[createPause]:")
	cli, err2 := GetNewClient()
	if err2 != nil {
		fmt.Println("error on creating Pause Container:")
		fmt.Println(err2)
		return container.ContainerCreateCreatedBody{}, err2
	}
	exports := make(nat.PortSet, len(ports))
	for _, port := range ports {
		//默认是tcp
		if port.Protocol == "" || port.Protocol == "tcp" || port.Protocol == "all" {
			p, err := nat.NewPort("tcp", port.Port)
			if err != nil {
				return container.ContainerCreateCreatedBody{}, err
			}
			exports[p] = struct{}{}
		}
		if port.Protocol == "udp" || port.Protocol == "all" {
			p, err := nat.NewPort("udp", port.Port)
			if err != nil {
				return container.ContainerCreateCreatedBody{}, err
			}
			exports[p] = struct{}{}
		}
	}
	resp, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        "registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.6",
			ExposedPorts: exports, //所有暴露出来的接口
		},
		&container.HostConfig{
			IpcMode: "shareable",
			DNS:     []string{ServiceDns}, //暂时在本文件设置一个const，以后可以写在config文件里
		},
		nil, nil, name)
	return resp, err
}

// 查找是否存在，存在就把原来的删除，之后统一创建新的。创建pod时用
func deleteExistedContainers(names []string) error {
	fmt.Println("[deleteExistedContainers]:")
	cli, err2 := GetNewClient()
	if err2 != nil {
		return err2
	}
	for _, value := range names {
		_, err := cli.ContainerInspect(context.Background(), value)
		if err == nil {
			//需要先停止container才能删除
			err = cli.ContainerStop(context.Background(), value, nil)
			// err = cli.ContainerStop(context.Background(), value, container.StopOptions{})
			if err == nil {
				//删除
				err = cli.ContainerRemove(context.Background(), value, types.ContainerRemoveOptions{})
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func checkImageExists(a string, tags []string) bool {
	for _, b := range tags {
		if a == b {
			return true
		}
		tmp := a + ":latest"
		if tmp == b {
			return true
		}
	}
	return false
}

// 拉取单个镜像
func dockerClientPullSingleImage(image string) error {
	fmt.Printf("[PullSingleImage] Prepare pull image:%s\n", image)
	cli, err2 := GetNewClient()
	if err2 != nil {
		return err2
	}
	out, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		fmt.Printf("[PullSingleImage] Fail to pull image, err:%v\n", err)
		return err
	}
	defer out.Close()
	io.Copy(ioutil.Discard, out)
	return nil
}

func dockerClientPullImages(images []string) error {
	fmt.Println("dockerClientPullImages:")
	//先统一拉取镜像，确认是否已经存在于本地
	cli, err2 := GetNewClient()
	if err2 != nil {
		return err2
	}
	resp, err := cli.ImageList(context.Background(), types.ImageListOptions{All: true})
	if err != nil {
		return err
	}
	var filter []string
	for _, value := range images {
		flag := false //此镜像是否已在本地
		for _, it := range resp {
			if checkImageExists(value, it.RepoTags) {
				fmt.Printf("[Kubelet] image %s exists \n", value)
				flag = true
				break
			}
		}
		if flag {
			continue
		}
		filter = append(filter, value)
	}
	// 剩下的是本地还不存在的，要单独拉取
	// if filter != nil {
	for _, value := range filter {
		err := dockerClientPullSingleImage(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func runContainers(containerIds []object.ContainerMeta) error {
	fmt.Println("[dockerClient] runContainers: ", containerIds)
	cli, err2 := GetNewClient()
	if err2 != nil {
		return err2
	}
	for _, value := range containerIds {
		fmt.Printf("[Kubelet] Run Container with ID %s \n", value)
		err := cli.ContainerStart(context.Background(), value.ContainerId, types.ContainerStartOptions{})
		if err != nil {
			fmt.Println("[Kubelet] start container fail", err)
			return err
		}
	}
	return nil
}

func getContainerNetInfo(name string) (*types.NetworkSettings, error) {
	cli, err1 := GetNewClient()
	if err1 != nil {
		fmt.Println(err1)
		return nil, err1
	}
	res, err := cli.ContainerInspect(context.Background(), name)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return res.NetworkSettings, nil
}

// 创建Pod的容器（先创建pause容器再创建要求的）
func createContainersOfPod(containers []object.Container) ([]object.ContainerMeta, *types.NetworkSettings, error) {
	fmt.Println("[dockerClient] createContainersOfPod")
	cli, err2 := client.NewClientWithOpts()
	if err2 != nil {
		return nil, nil, err2
	}
	var firstContainerId string
	var result []object.ContainerMeta
	//暴露的端口集合
	var totalPort []object.ContainerPort
	images := []string{"registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.6"}
	//如果有重名要先删除
	var names []string
	pauseName := "pause"
	for _, value := range containers {
		//pause容器名字附带当前容器的名字
		pauseName += "_" + value.Name
		//名字列表加入当前容器的名字
		names = append(names, value.Name)
		//镜像列表加上当前镜像
		images = append(images, value.Image)
		for _, port := range value.Ports {
			//添加到暴露的所有端口中
			totalPort = append(totalPort, port)
		}
	}
	names = append(names, pauseName)
	//先将列表中之前存在的容器删掉，之后再统一启动（）
	err3 := deleteExistedContainers(names)
	if err3 != nil {
		return nil, nil, err3
	}
	//拉取所有镜像（先把本地的拉取，再分别单个拉取不在本地的镜像）
	err := dockerClientPullImages(images)
	if err != nil {
		return nil, nil, err
	}
	//创建pause容器
	pause, err := createPause(totalPort, pauseName)
	fmt.Println("[Kubelet] pausename:", pauseName)
	if err != nil {
		return nil, nil, err
	}
	firstContainerId = pause.ID
	result = append(result, object.ContainerMeta{
		RealName:    pauseName,
		ContainerId: firstContainerId,
	})
	var tmpContainers []container.ContainerCreateCreatedBody
	tmpContainers = append(tmpContainers, pause)
	for _, value := range containers {
		fmt.Println("[Kubelet] containerName:", value.Name)
		//fmt.Println("[Kubelet] commandTest:", value.Command)
		//fmt.Println("[Kubelet] argTest:", value.Args)
		var mounts []mount.Mount
		if value.VolumeMounts != nil {
			for _, it := range value.VolumeMounts {
				mounts = append(mounts, mount.Mount{
					Type:   mount.TypeBind,
					Source: it.Name,
					Target: it.MountPath,
				})
			}
		}
		//生成env
		var env []string
		if value.Env != nil {
			for _, it := range value.Env {
				singleEnv := it.Name + "=" + it.Value
				env = append(env, singleEnv)
			}
		}
		//resource
		resourceConfig := container.Resources{}
		//创建容器
		fmt.Printf("[dockerClient] ContainerCreate with image %s \n", value.Image)
		resp, err := cli.ContainerCreate(
			context.Background(),
			&container.Config{
				Image:      value.Image,
				Entrypoint: value.Command,
				Cmd:        value.Args,
				Env:        env,
			},
			&container.HostConfig{
				NetworkMode: container.NetworkMode("container:" + firstContainerId),
				Mounts:      mounts,
				IpcMode:     container.IpcMode("container:" + firstContainerId),
				PidMode:     container.PidMode("container" + firstContainerId),
				Resources:   resourceConfig,
			},
			nil, nil, value.Name)
		if err != nil {
			fmt.Println("[Kubelet] run container fail with reason", err)
			return nil, nil, err
		}
		tmpContainers = append(tmpContainers, resp)
		// //获取信息
		// getContainerInfo(resp.ID)
		// //
		result = append(result, object.ContainerMeta{
			RealName:    value.Name,
			ContainerId: resp.ID,
		})
	}
	//-----------获取所有容器信息---------
	// fmt.Println("Show Containers Info:-----")
	// for _, value := range tmpContainers {
	// 	getContainerInfo(value.ID)
	// }
	//-----------获取所有容器信息---------

	//启动容器
	err = runContainers(result)
	if err != nil {
		return nil, nil, err
	}
	var netSetting *types.NetworkSettings
	netSetting, err = getContainerNetInfo(pauseName)
	if err != nil {
		return nil, nil, err
	}
	return result, netSetting, nil
}

// 删除指定容器
func deleteContainers(containerIds []string) error {
	cli, err2 := GetNewClient()
	if err2 != nil {
		return err2
	}
	//需要先停止containers
	for _, value := range containerIds {
		err := cli.ContainerStop(context.Background(), value, nil)
		if err != nil {
			return err
		}
	}
	//停止后删除
	for _, value := range containerIds {
		err := cli.ContainerRemove(context.Background(), value, types.ContainerRemoveOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// 探针检查容器状态
func probeContainers(containerIds []string) ([]string, error) {
	cli, err2 := GetNewClient()
	if err2 != nil {
		fmt.Println(err2)
		return nil, err2
	}
	var res []string
	for _, value := range containerIds {
		resp, err := cli.ContainerInspect(context.Background(), value)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		res = append(res, resp.State.Status)
	}
	return res, nil
}

func Main(Group []object.Container) {
	//p := (*message.CommandWithConfig)(unsafe.Pointer(command))
	//Group := []object.Container{}
	//res, netSetting, err := createContainersOfPod(p.Group)
	res, netSetting, err := createContainersOfPod(Group)
	if res == nil || netSetting == nil || err != nil {
		fmt.Println(".......")
	}
	//var result message.ResponseWithContainIds
	//result.Err = err
	//result.CommandType = message.COMMAND_BUILD_CONTAINERS_OF_POD
	//result.Containers = res
	//result.NetWorkInfos = netSetting
}

func HandleCommand(command *message.Command) *message.Response {
	switch command.CommandType {
	case message.COMMAND_BUILD_CONTAINERS_OF_POD:
		//创建pod的容器
		p := (*message.CommandWithConfig)(unsafe.Pointer(command))
		res, netSetting, err := createContainersOfPod(p.Group)
		var result message.ResponseWithContainIds
		result.Err = err
		result.CommandType = message.COMMAND_BUILD_CONTAINERS_OF_POD
		result.Containers = res
		result.NetWorkInfos = netSetting
		return &(result.Response)
	case message.COMMAND_DELETE_CONTAINER:
		//删除containers
		p := (*message.CommandWithContainerIds)(unsafe.Pointer(command))
		err := deleteContainers(p.ContainerIds)
		var result message.Response
		result.CommandType = message.COMMAND_DELETE_CONTAINER
		result.Err = err
		return &result
	case message.COMMAND_PROBE_CONTAINER:
		//liveness probe 查看是否存活
		p := (*message.CommandWithContainerIds)(unsafe.Pointer(command))
		res, err := probeContainers(p.ContainerIds)
		var result message.ResponseWithProbeInfos
		result.Err = err
		result.CommandType = message.COMMAND_PROBE_CONTAINER
		result.ProbeInfos = res
		return &(result.Response)
	}
	return nil
}
