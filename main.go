package main

import "github.com/gmtstephane/kpture/cmd"

func main() {
	cmd.Execute()

	// logfile, err := os.OpenFile("kpture.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	// if err != nil {
	// 	logrus.Error(err)
	// 	return
	// }
	// logrus.SetOutput(logfile)

	// logrus.SetFormatter(&nested.Formatter{
	// 	HideKeys:       true,
	// 	NoColors:       true,
	// 	NoFieldsColors: true,
	// 	FieldsOrder:    []string{"component", "category"},
	// })
	// // pods := []*kpture.KpturePod{}

	// client, err := kpture.GetClient()
	// if err != nil {
	// 	logrus.Error(err)
	// 	return
	// }

	// podList := []kpture.PodDescriptor{}

	// pods, err := client.Clientset.List(context.Background(), v1.ListOptions{})
	// if err != nil {
	// 	logrus.Error(err)
	// 	return
	// }
	// for _, pod := range pods.Items {
	// 	podList = append(podList, kpture.PodDescriptor{
	// 		Name:      pod.Name,
	// 		Namespace: pod.Namespace,
	// 	})
	// }

	// kpture, err := kpture.NewKpture(client, podList)
	// if err != nil {
	// 	logrus.Error(err)
	// 	return
	// }

	// c := make(chan os.Signal)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// go func() {
	// 	<-c
	// 	kpture.Stop()
	// 	os.Exit(1)
	// }()

	// err = kpture.SetupEphemeralContainers()
	// if err != nil {
	// 	logrus.Error(err)
	// 	return
	// }

	// err = kpture.SetupPortForwarding()
	// if err != nil {
	// 	logrus.Error(err)
	// 	return
	// }

	// kpture.ReadPacketsConn()

	// err = kpture.HandlePackets(os.Stdout)
	// if err != nil {
	// 	logrus.Error(err)
	// 	return
	// }
}
