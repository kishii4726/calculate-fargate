package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

func getEcsClusters(c *ecs.Client) []string {
	resp, err := c.ListClusters(context.TODO(), &ecs.ListClustersInput{
		MaxResults: aws.Int32(100),
	})
	if err != nil {
		log.Fatalf("ListClusters failed %v\n", err)
	}
	ecs_cluster_arns := resp.ClusterArns
	if len(ecs_cluster_arns) == 0 {
		log.Fatalf("Cluster does not exist")
	}
	ecs_clusters := []string{}
	for _, v := range ecs_cluster_arns {
		ecs_clusters = append(ecs_clusters, strings.Split(v, "/")[1])
	}

	return ecs_clusters
}

func getEcsServices(client *ecs.Client, ecs_cluster string) []string {
	resp, err := client.ListServices(context.TODO(), &ecs.ListServicesInput{
		Cluster:    aws.String(ecs_cluster),
		MaxResults: aws.Int32(100),
	})
	if err != nil {
		log.Fatalf("ListServices failed %v\n", err)
	}

	return resp.ServiceArns
}

func getEcsTaskIds(client *ecs.Client, ecs_cluster string, ecs_service string) []string {
	resp, err := client.ListTasks(context.TODO(), &ecs.ListTasksInput{
		Cluster:     aws.String(ecs_cluster),
		ServiceName: aws.String(ecs_service),
		MaxResults:  aws.Int32(100),
	})
	if err != nil {
		log.Fatalf("ListTasks failed %v\n", err)
	}
	ecs_task_arns := resp.TaskArns
	ecs_task_ids := []string{}
	if len(ecs_task_arns) != 0 {
		for _, v := range ecs_task_arns {
			ecs_task_ids = append(ecs_task_ids, strings.Split(v, "/")[2])
		}
	}

	return ecs_task_ids
}

func getContainers(client *ecs.Client, ecs_cluster string, ecs_task_id string) []string {
	resp, err := client.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
		Tasks:   []string{ecs_task_id},
		Cluster: aws.String(ecs_cluster),
	})
	if err != nil {
		log.Fatalf("DescribeTasks failed %v\n", err)
	}
	ecs_containers := resp.Tasks[0].Containers
	ecs_container_names := []string{}
	for _, v := range ecs_containers {
		ecs_container_names = append(ecs_container_names, *v.Name)
	}
	return ecs_container_names
}

func main() {

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	client := ecs.NewFromConfig(cfg)
	var total_tasks_in_the_account int
	var total_containers_in_the_account int
	for _, cluster := range getEcsClusters(client) {
		number_of_tasks_in_cluster := []string{}
		fmt.Println("Cluster name: " + cluster)
		for _, service_arn := range getEcsServices(client, cluster) {
			for _, task_id := range getEcsTaskIds(client, cluster, service_arn) {
				number_of_tasks_in_cluster = append(number_of_tasks_in_cluster, task_id)
				total_containers_in_the_account += len(getContainers(client, cluster, task_id))
			}
		}
		total_tasks_in_the_account += len(number_of_tasks_in_cluster)
		fmt.Println("Number of tasks in cluster: " + strconv.Itoa(len(number_of_tasks_in_cluster)))
	}

	fmt.Println("")
	fmt.Println("Total number of tasks in account: " + strconv.Itoa(total_tasks_in_the_account))
	fmt.Println("Total number of containers in account: " + strconv.Itoa(total_containers_in_the_account))
}
