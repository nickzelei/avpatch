package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	fmt.Println(os.Args)
	if len(os.Args) != 3 {
		panic(fmt.Errorf("must provide additional args"))
	}
	userName := os.Args[1]
	awsProfile := os.Args[2]

	node, err := getKubeConfig()
	if err != nil {
		panic(err)
	}
	root := node.Content[0]
	usersNode := valueOf(root, "users")
	for _, userNode := range usersNode.Content {
		nameNode := valueOf(userNode, "name")
		if nameNode.Kind == yaml.ScalarNode && nameNode.Value == userName {
			usrNode := valueOf(userNode, "user")
			if usrNode.Kind == yaml.MappingNode {
				execNode := valueOf(usrNode, "exec")
				if execNode.Kind == yaml.MappingNode {
					cmdNode := valueOf(execNode, "command")
					if cmdNode.Kind == yaml.ScalarNode && cmdNode.Value == "aws" {
						cmdNode.Value = "aws-vault"
						argsNode := valueOf(execNode, "args")
						if argsNode.Kind == yaml.SequenceNode {
							argsNode.Content = append([]*yaml.Node{
								{
									Kind:  yaml.ScalarNode,
									Value: "exec",
									Tag:   "!!str",
								},
								{
									Kind:  yaml.ScalarNode,
									Value: awsProfile,
									Tag:   "!!str",
								},
								{
									Kind:  yaml.ScalarNode,
									Value: "--",
									Tag:   "!!str",
								},
							}, argsNode.Content...)
						}
					}
				}
			}
		}
	}

	kpath, err := getKubeconfigPath()
	if err != nil {
		panic(err)
	}
	f, err := os.OpenFile(kpath, os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	enc := yaml.NewEncoder(f)
	enc.SetIndent(0)
	err = enc.Encode(root)
	if err != nil {
		panic(err)
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}
}

func getKubeConfig() (*yaml.Node, error) {
	kpath, err := getKubeconfigPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(kpath)
	if err != nil {
		return nil, err
	}

	var v yaml.Node
	err = yaml.NewDecoder(f).Decode(&v)
	if err != nil {
		return nil, err
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func getKubeconfigPath() (string, error) {
	// KUBECONFIG env var
	if v := os.Getenv("KUBECONFIG"); v != "" {
		list := filepath.SplitList(v)
		if len(list) > 1 {
			// TODO KUBECONFIG=file1:file2 currently not supported
			return "", errors.New("multiple files in KUBECONFIG are currently not supported")
		}
		return v, nil
	}

	// default path
	home := getHomeDir()
	if home == "" {
		return "", errors.New("HOME or USERPROFILE environment variable not set")
	}
	return filepath.Join(home, ".kube", "config"), nil
}

func getHomeDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // windows
	}
	return home
}

func valueOf(mapNode *yaml.Node, key string) *yaml.Node {
	if mapNode.Kind != yaml.MappingNode {
		return nil
	}
	for i, ch := range mapNode.Content {
		if i%2 == 0 && ch.Kind == yaml.ScalarNode && ch.Value == key {
			return mapNode.Content[i+1]
		}
	}
	return nil
}
