package authz

//var UserBindingMap = hashmap.New()
//var GroupBindingMap = hashmap.New()
//var GroupUserMap = hashmap.New()
//var RoleBindingExtAllowMap = hashmap.New()
//var RoleBindingExtDenyMap = hashmap.New()
//var ClientSet *menshenext.MenshenV1Beta1Client
//
//type RoleBindingExtInfo struct {
//	RoleExtNames 	[]string
//	Message			string
//}

func Cache() {
	//kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	//// BuildConfigFromFlags is a helper function that builds configs from a master url or
	//// a kubeconfig filepath.
	//config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	//utils.HandleErr(err)
	//
	//ClientSet, err = menshenext.NewForConfig(config)
	//utils.HandleErr(err)

	//roleExtList, err := ClientSet.RoleExts().List(metav1.ListOptions{})
	//utils.HandleErr(err)
	//
	//for _, roleExt := range roleExtList.Items {
	//	roleName := roleExt.GetName()
	//	roleMap := menshencontroller.GetRoleExtInfo(roleExt.Spec.Roles)
	//	RoleExtMap.Put(roleName, roleMap)
	//}

	//groupextList := &menshenv1beta1.GroupExtList{}
	//groupextList, err = ClientSet.GroupExts().List(metav1.ListOptions{})
	//utils.HandleErr(err)
	//for _, groupExt := range groupextList.Items {
	//	GroupUserMap.Put(groupExt.Name, groupExt.Spec.Users)
	//}

	//rolebindingextList, err := ClientSet.RoleBindingExts().List(metav1.ListOptions{})
	//utils.HandleErr(err)
	//for _, rolebindingext := range rolebindingextList.Items {
	//
	//	for _, subject := range rolebindingext.Spec.Subjects {
	//		if strings.EqualFold(subject.Kind, "User") ||
	//			strings.EqualFold(subject.Kind, "ServiceAccount") {
	//
	//			if rolebindingExtNameList, found := UserBindingMap.Get(subject.Name); found {
	//				rolebindingExtNameList = append(rolebindingExtNameList.([]string), rolebindingext.Name)
	//			} else {
	//				rolebindingExtNameList := []string{}
	//				rolebindingExtNameList = append(rolebindingExtNameList, rolebindingext.Name)
	//				UserBindingMap.Put(subject.Name, rolebindingExtNameList)
	//			}
	//		} else if strings.EqualFold(subject.Kind, "Group") {
	//
	//			if rolebindingExtNameList, found := GroupBindingMap.Get(subject.Name); found {
	//				rolebindingExtNameList = append(rolebindingExtNameList.([]string), rolebindingext.Name)
	//			} else {
	//				rolebindingExtNameList := []string{}
	//				rolebindingExtNameList = append(rolebindingExtNameList, rolebindingext.Name)
	//				GroupBindingMap.Put(subject.Name, rolebindingExtNameList)
	//			}
	//		}
	//	}
	//
	//	if strings.EqualFold(rolebindingext.Spec.Type, "allow") {
	//		RoleBindingExtAllowMap.Put(rolebindingext.Name, RoleBindingExtInfo{RoleExtNames: rolebindingext.Spec.RoleNames})
	//	}else if strings.EqualFold(rolebindingext.Spec.Type, "deny") {
	//		RoleBindingExtDenyMap.Put(rolebindingext.Name, RoleBindingExtInfo{RoleExtNames: rolebindingext.Spec.RoleNames, Message: rolebindingext.Spec.Message})
	//	}else {
	//		// TODO
	//		klog.Error("Without type:" + rolebindingext.Name)
	//	}
	//}
}
