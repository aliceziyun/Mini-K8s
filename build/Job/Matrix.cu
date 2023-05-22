#include <cuda_runtime.h>
#include "device_launch_parameters.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

// 30 * 30的矩阵

// 随机初始化数组
void initialInt(float *ip, float size)
{
	for (int i = 0; i < size; i++)
	{
		ip[i] = (float)(rand() & 0xff) / 66.6;
	}
}
// 打印数组
void printMatrix(float *A, float *B, float *C, const int nx, const int ny)
{
	float *ia = A, *ib = B, *ic = C;
	printf("\nMatrix:(%d, %d)\n", nx, ny);
	for (int iy = 0; iy < ny; iy++)
	{
		for (int ix = 0; ix < nx; ix++)
		{
			printf("%f + %f = %f     ", ia[ix], ib[ix], ic[ix]);
		}
		ia += nx;
		ib += nx;
		ic += nx;
		printf("\n");
	}
	printf("\n");
}
// 验证结果
void printResult(float *C, float *CC, const int nx, const int ny)
{
	float *ic = C, *icc = CC;
	for (int iy = 0; iy < ny; iy++)
	{
		for (int ix = 0; ix < nx; ix++)
		{
			printf("%f     ", ic[ix]-icc[ix]);
		}
		ic += nx;
		icc += nx;
		printf("\n");
	}
	printf("\n");
}

// GPU：计算C=A+B
__global__ void sumMatrixOnDevice(float *MatA, float *MatB, float *MatC, const int nx, const int ny)
{
	int ix = threadIdx.x + blockDim.x*blockIdx.x;
	int iy = threadIdx.y + blockDim.y*blockIdx.y;
	unsigned int idx = iy * nx + ix;
	//unsigned int t_n = gridDim.x*blockDim.x + gridDim.y*blockDim.y;
	if (ix < nx && iy < ny)
	{
		MatC[idx] = MatA[idx] + MatB[idx];
	}
}

// GPU：计算C=A*B
__global__ void MatMul(float *M,float *N,float *P,int width)
{
	int x = threadIdx.x;
	int y = threadIdx.y;

	float Pervalue = 0;

	float elem1 = 0.0,elem2 = 0.0,value = 0.0;
	for(int i = 0;i < width;i++)
	{
		elem1 = M[y * width + i];//取M矩阵的一行
		elem2 = N[i * width + x];//取N矩阵的一列

		value += elem1 * elem2;//求和
	}

	P[y * width + x] = value;
}


int main(int argc, char **argv)
{
	//printf("%s Starting...\n", argv[10]);

	int dev = 0;
	cudaDeviceProp deviceProp;
	cudaGetDeviceProperties(&deviceProp, dev);
	printf("Using Device  %d: %s\n\n", dev, deviceProp.name);

	// set matrix dimension
	int nx = 30;
	int ny = 30;
	int nxy = nx * ny;
	int nBytes = nxy * sizeof(float);

	// malloc host dimension
	float *h_A, *h_B, *h_C, *h_CC;
	h_A = (float *)malloc(nBytes);
	h_B = (float *)malloc(nBytes);
	h_C = (float *)malloc(nBytes);
	h_CC = (float *)malloc(nBytes);

	// initialize host matrix with integer
	initialInt(h_A, nxy);
	initialInt(h_B, nxy);

	// 开始计时
	clock_t cpuStart = clock();

	sumMatrixOnHost(h_A, h_B, h_C, nx, ny);

	// 结束计时
	clock_t cpuEnd = clock();
	float cpuTime = (float)(cpuEnd - cpuStart) / CLOCKS_PER_SEC;
	printf("cpu time:%f\n", cpuTime);

	// mallox device memory
	float *d_MatA, *d_MatB, *d_MatC;
	cudaMalloc((void **)&d_MatA, nBytes);
	cudaMalloc((void **)&d_MatB, nBytes);
	cudaMalloc((void **)&d_MatC, nBytes);

	// 开始计时
// 	clock_t gpuStart = clock();

	// transfer data from host to device
	cudaMemcpy(d_MatA, h_A, nBytes, cudaMemcpyHostToDevice);
	cudaMemcpy(d_MatB, h_B, nBytes, cudaMemcpyHostToDevice);

	//set up execution configuration
	int dimx = 32;
	int dimy = 32;
	dim3 block(dimx, dimy);
	dim3 grid((nx + block.x - 1) / block.x, (ny + block.y - 1) / block.y);


	// 矩阵加法
	sumMatrixOnDevice << <grid, block >> > (d_MatA, d_MatB, d_MatC, nx, ny);
	cudaDeviceSynchronize();
	// transfer data from device to host
	cudaMemcpy(h_CC, d_MatC, nBytes, cudaMemcpyDeviceToHost);
	printResult(h_C, h_CC, nx, ny);

	// 矩阵乘法
	MatMul<<<1,blockSize>>>(d_MatA,d_MatB,d_MatC,nx);//调用核函数
	cudaThreadSynchronize();
	cudaMemcpy(h_CC,d_MatC,nBytes,cudaMemcpyDeviceToHost);
    printf("c0 = %d \n",h_CC[0][0]);


	// 结束计时
    // 	clock_t gpuEnd = clock();
    // 	float gpuTime = (float)(gpuEnd - gpuStart) / CLOCKS_PER_SEC;
    // 	printf("gpu time:%f\n", gpuTime);

	// free host and device memory
	cudaFree(d_MatA);
	cudaFree(d_MatB);
	cudaFree(d_MatC);
	free(h_A);
	free(h_B);
	free(h_C);

	// reset device
	cudaDeviceReset();

	return 0;
}