//有些暂时用来参考
package dockerClient

import (
	"Mini-K8s/pkg/object"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const ServiceDns string = "10.10.10.10"

func GetNewClient() (*client.Client, error) {
	return client.NewClientWithOpts()
}

func startContainer(containerId string) error {
	cli, err := GetNewClient()
	//cli, err := client.NewClientWithOpts()
	if err != nil {
		// fmt.Print("%v\n", err)
		return err
	}
	err = cli.ContainerStart(context.Background(), containerId, types.ContainerStartOptions{})
	return err
}

//创建pause容器
func createPause(ports []object.ContainerPort, name string) (container.ContainerCreateCreatedBody, error) {
	fmt.Println("createPause:")
	cli, err2 := GetNewClient()
	//cli, err2 := client.NewClientWithOpts()
	if err2 != nil {
		return container.ContainerCreateCreatedBody{}, err2
	}
	var exports nat.PortSet
	exports = make(nat.PortSet, len(ports))
	for _, port := range ports {
		if port.Protocol == "" || port.Protocol == "tcp" || port.Protocol == "all" {
			p, err := nat.NewPort("tcp", port.ContainerPort)
			if err != nil {
				return container.ContainerCreateCreatedBody{}, err
			}
			exports[p] = struct{}{}
		}
		if port.Protocol == "udp" || port.Protocol == "all" {
			p, err := nat.NewPort("udp", port.ContainerPort)
			if err != nil {
				return container.ContainerCreateCreatedBody{}, err
			}
			exports[p] = struct{}{}
		}
	}
	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image:        "registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.6",
		ExposedPorts: exports,
	}, &container.HostConfig{
		IpcMode: container.IpcMode("shareable"),
		//DNS:     []string{netconfig.ServiceDns},
		DNS: []string{ServiceDns}, //暂时在本文件设置一个const，以后可以写在config文件里
		//const ServiceDns = "10.10.10.10"
	}, nil, nil, name)
	return resp, err
}

//查找是否存在，存在就删除
func deleteExitedContainers(names []string) error {
	fmt.Println("deleteExitedContainers:")
	cli, err2 := GetNewClient()
	if err2 != nil {
		return err2
	}
	for _, value := range names {
		_, err := cli.ContainerInspect(context.Background(), value)
		if err == nil {
			//需要先停止container
			//err = cli.ContainerStop(context.Background(), value, nil)
			err = cli.ContainerStop(context.Background(), value, container.StopOptions{})
			if err != nil {
				return err
			}
			err = cli.ContainerRemove(context.Background(), value, types.ContainerRemoveOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func isImageExist(a string, tags []string) bool {
	for _, b := range tags {
		if a == b {
			return true
		}
		tmp := a + ":latest"
		if tmp == b {
			return true
		}
	}

	fmt.Printf("Local image:%v Target image:%s\n", tags, a)
	return false
}

//注意， 调用ImagePull 函数， 拉取进程在后台运行，因此要保证前台挂起足够时间保证拉取成功
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
		flag := false
		for _, it := range resp {
			if isImageExist(value, it.RepoTags) {
				flag = true
				break
			}
		}
		if flag {
			continue
		}
		filter = append(filter, value)
	}
	if filter != nil {
		for _, value := range filter {
			err := dockerClientPullSingleImage(value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
func runContainers(containerIds []object.ContainerMeta) error {
	fmt.Println("runContainers:")
	cli, err2 := GetNewClient()
	if err2 != nil {
		return err2
	}
	for _, value := range containerIds {
		err := cli.ContainerStart(context.Background(), value.ContainerId, types.ContainerStartOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
func getContainerNetInfo(name string) (*types.NetworkSettings, error) {
	//cli, err2 := getNewClient()
	//if err2 != nil {
	//	return nil, err2
	//}
	//res, err := cli.ContainerInspect(context.Background(), name)
	//if err != nil {
	//	return nil, err
	//}
	//return res.NetworkSettings, nil
	return nil, nil
}
func createContainersOfPod(containers []object.Container) ([]object.ContainerMeta, *types.NetworkSettings, error) {
	fmt.Println("createContainersOfPod:")
	//cli, err2 := getNewClient()
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
		pauseName += "_" + value.Name
		names = append(names, value.Name)
		images = append(images, value.Image)
		for _, port := range value.Ports {
			totalPort = append(totalPort, port)
		}
	}
	names = append(names, pauseName)
	err3 := deleteExitedContainers(names)
	if err3 != nil {
		return nil, nil, err3
	}
	//拉取所有镜像（分别拉取单个镜像）
	err := dockerClientPullImages(images)
	if err != nil {
		return nil, nil, err
	}
	//创建pause容器
	pause, err := createPause(totalPort, pauseName)
	if err != nil {
		return nil, nil, err
	}
	firstContainerId = pause.ID
	result = append(result, object.ContainerMeta{
		RealName:    pauseName,
		ContainerId: firstContainerId,
	})
	for _, value := range containers {
		//var mounts []mount.Mount
		//if value.VolumeMounts != nil {
		//	for _, it := range value.VolumeMounts {
		//		mounts = append(mounts, mount.Mount{
		//			Type:   mount.TypeBind,
		//			Source: it.Name,
		//			Target: it.MountPath,
		//		})
		//	}
		//}
		//生成env
		var env []string
		if value.Env != nil {
			for _, it := range value.Env {
				singleEnv := it.Name + "=" + it.Value
				env = append(env, singleEnv)
			}
		}
		//生成resource
		resourceConfig := container.Resources{}
		//if value.Limits.Cpu != "" {
		//	resourceConfig.NanoCPUs = getCpu(value.Limits.Cpu)
		//}
		//if value.Limits.Memory != "" {
		//	resourceConfig.Memory = getMemory(value.Limits.Memory)
		//}
		fmt.Println("ContainerCreate")
		resp, err := cli.ContainerCreate(context.Background(), &container.Config{
			Image:      value.Image,
			Entrypoint: value.Command,
			Cmd:        value.Args,
			Env:        env,
		}, &container.HostConfig{
			NetworkMode: container.NetworkMode("container:" + firstContainerId),
			//Mounts:      mounts,
			IpcMode:   container.IpcMode("container:" + firstContainerId),
			PidMode:   container.PidMode("container" + firstContainerId),
			Resources: resourceConfig,
		}, nil, nil, value.Name)
		if err != nil {
			return nil, nil, err
		}
		result = append(result, object.ContainerMeta{
			RealName:    value.Name,
			ContainerId: resp.ID,
		})
	}
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

func Main(Group []object.Container) {
	//p := (*message.CommandWithConfig)(unsafe.Pointer(command))
	//Group := []object.Container{}
	//res, netSetting, err := createContainersOfPod(p.Group)
	res, netSetting, err := createContainersOfPod(Group)
	if res == nil || netSetting == nil || err == nil {

	}
	//var result message.ResponseWithContainIds
	//result.Err = err
	//result.CommandType = message.COMMAND_BUILD_CONTAINERS_OF_POD
	//result.Containers = res
	//result.NetWorkInfos = netSetting
}
