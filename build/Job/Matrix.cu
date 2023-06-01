#include <stdio.h>
#include <stdlib.h>
#include <time.h>

#include <cuda_runtime.h>

#define THREAD_NUM 256

#define MATRIX_SIZE 10

const int blocks_num = (MATRIX_SIZE + THREAD_NUM - 1) / THREAD_NUM;

// CUDA 初始化
bool InitCUDA()
{
    int count;
    cudaGetDeviceCount(&count);
    if (count == 0)
    {
        fprintf(stderr, "There is no device.\n");
        return false;
    }

    int i;
    for (i = 0; i < count; i++)
    {
        cudaDeviceProp prop;
        cudaGetDeviceProperties(&prop, i);
        if (cudaGetDeviceProperties(&prop, i) == cudaSuccess)
        {
            if (prop.major >= 1)
            {
                break;
            }
        }
    }

    if (i == count)
    {
        fprintf(stderr, "There is no device supporting CUDA 1.x.\n");
        return false;
    }

    cudaSetDevice(i);
    return true;
}

void generateMatrix(int *a, int size)
{
    for (int i = 0; i < size; i++)
    {
        for (int j = 0; j < size; j++)
        {
            a[i * size + j] = rand() % 256;
        }
    }
}

void printMatrix(int *a, int size)
{
    //print
    // puts("===========Print a Matrix===========");
    for (int i = 0; i < size; i++)
    {
        for (int j = 0; j < size; j++)
        {
            printf("%d ",a[i * size + j]);
        }
        puts("");
    }
    puts("");
}

__global__ static void addMatrixCUDA(const int *a, const int *b, int *c, int size)
{
    const int tid = threadIdx.x;
    const int bid = blockIdx.x;

    const int idx = bid * THREAD_NUM + tid;

    if (idx < size)
    {
        int max = size * size;
        for (int i = idx; i < max; i += size) {
            c[i] = a[i] + b[i];
        }
    }
}

__global__ static void multiMatrixCUDA(const int *a, const int *b, int *c, int size)
{
    const int tid = threadIdx.x;
    const int bid = blockIdx.x;

    const int idx = bid * THREAD_NUM + tid;
    const int row = idx / size;
    const int column = idx % size;

    if (row < size && column < size)
    {
        int t = 0;

        for (int i = 0; i < size; i++)
        {
            t += a[row * size + i] * b[i * size + column];
        }
        c[row * size + column] = t;
    }
}

int main()
{
    if (!InitCUDA())
        return 0;

    srand(0);

    int *a, *b, *c, *d;
    a = (int *)malloc(sizeof(int) * MATRIX_SIZE * MATRIX_SIZE);
    b = (int *)malloc(sizeof(int) * MATRIX_SIZE * MATRIX_SIZE);
    c = (int *)malloc(sizeof(int) * MATRIX_SIZE * MATRIX_SIZE);
    d = (int *)malloc(sizeof(int) * MATRIX_SIZE * MATRIX_SIZE);

    generateMatrix(a, MATRIX_SIZE);
    generateMatrix(b, MATRIX_SIZE);

    //print a, b
    puts("[Matrix a]:");
    printMatrix(a, MATRIX_SIZE);
    puts("[Matrix b]:");
    printMatrix(b, MATRIX_SIZE);

    int *cuda_a, *cuda_b, *cuda_c, *cuda_d;

    cudaMalloc((void **)&cuda_a, sizeof(int) * MATRIX_SIZE * MATRIX_SIZE);
    cudaMalloc((void **)&cuda_b, sizeof(int) * MATRIX_SIZE * MATRIX_SIZE);
    cudaMalloc((void **)&cuda_c, sizeof(int) * MATRIX_SIZE * MATRIX_SIZE);
    cudaMalloc((void **)&cuda_d, sizeof(int) * MATRIX_SIZE * MATRIX_SIZE);

    cudaMemcpy(cuda_a, a, sizeof(int) * MATRIX_SIZE * MATRIX_SIZE, cudaMemcpyHostToDevice);
    cudaMemcpy(cuda_b, b, sizeof(int) * MATRIX_SIZE * MATRIX_SIZE, cudaMemcpyHostToDevice);

    //加法
    addMatrixCUDA <<<blocks_num, THREAD_NUM, 0>>>(cuda_a, cuda_b, cuda_c, MATRIX_SIZE);
    cudaMemcpy(c, cuda_c, sizeof(int) * MATRIX_SIZE * MATRIX_SIZE, cudaMemcpyDeviceToHost);
    //print c
    puts("[Matrix c]:");
    printMatrix(c, MATRIX_SIZE);

    //乘法
    multiMatrixCUDA <<<blocks_num, THREAD_NUM, 0>>>(cuda_a, cuda_b, cuda_d, MATRIX_SIZE);
    cudaMemcpy(d, cuda_d, sizeof(int) * MATRIX_SIZE * MATRIX_SIZE, cudaMemcpyDeviceToHost);
    //print d
    puts("[Matrix d]:");
    printMatrix(d, MATRIX_SIZE);

    cudaFree(cuda_a);
    cudaFree(cuda_b);
    cudaFree(cuda_c);
    cudaFree(cuda_d);

    return 0;
}