package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	v1 "k8s.io/api/core/v1"
	extender "k8s.io/kube-scheduler/extender/v1"
)

const (
	// LuckyPred rejects a node if you're not lucky ¯\_(ツ)_/¯
	LuckyPred        = "Lucky"
	LuckyPredFailMsg = "Well, you're not lucky"
)

var predicatesFuncs = map[string]FitPredicate{
	LuckyPred: LuckyPredicate,
}

type FitPredicate func(pod *v1.Pod, node v1.Node) (bool, []string, error)

var predicatesSorted = []string{LuckyPred}

// filter filters nodes according to predicates defined in this extender
// it's webhooked to pkg/scheduler/core/generic_scheduler.go#findNodesThatFitPod()
func filter(args extender.ExtenderArgs) *extender.ExtenderFilterResult {
	var filteredNodes []v1.Node

	failedNodes := make(extender.FailedNodesMap)
	pod := args.Pod
	// create a 2 d array of nodes
	// nodes names will be the key
	// watt value will be the value
	var wattnodemap = make(map[string]int)

	for _, node := range args.Nodes.Items {

		wattnodemap[node.Name] = rand.Intn(2)
	}
	// Creating and initializing a map
	// Using shorthand declaration and
	// using map literals
	regionWatt := map[string]int{

		"MISO_MI": 78,
		"ISONE_WCMA	": 7,
		"PJM_ATLANTIC": 0,
		"SOCO":         27,
		"AECI":         8,
	}

	fmt.Println("Node's watt value: ", regionWatt["USA"])
	value1 := regionWatt["USA"]
	// TODO: parallelize this
	// TODO: handle error
	for _, node := range args.Nodes.Items {
		fits, failReasons, _ := podFitsOnNode(pod, node)
		if fits {
			filteredNodes = append(filteredNodes, node)
		} else {
			failedNodes[node.Name] = strings.Join(failReasons, ",")
		}
	}

	result := extender.ExtenderFilterResult{
		Nodes: &v1.NodeList{
			Items: filteredNodes,
		},
		FailedNodes: failedNodes,
		Error:       "",
	}

	return &result
}

// This is the method which should be changed for Watt API
func podFitsOnNode(pod *v1.Pod, node v1.Node) (bool, []string, error) {
	fits := true
	var failReasons []string
	for _, predicateKey := range predicatesSorted {
		fit, failures, err := predicatesFuncs[predicateKey](pod, node)
		if err != nil {
			return false, nil, err
		}
		fits = fits && fit
		failReasons = append(failReasons, failures...)
	}
	return fits, failReasons, nil
}

//Luckypredicate  main function where we will
//Build array of regions and their corresponding WATT API values
// Lucky predicate will choose the lowest value and return.
func LuckyPredicate(pod *v1.Pod, node v1.Node) (bool, []string, error) {

	lucky := rand.Intn(2) == 0
	if lucky {
		log.Printf("pod %v/%v is lucky to fit on node %v\n", pod.Name, pod.Namespace, node.Name)
		return true, nil, nil
	}
	log.Printf("pod %v/%v is unlucky to fit on node %v\n", pod.Name, pod.Namespace, node.Name)
	return false, []string{LuckyPredFailMsg}, nil
}
