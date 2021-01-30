package e2e

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedRbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

func createCMEditRoleAndBinding(ctx context.Context, rbacv1Client typedRbacv1.RbacV1Interface, ns string) error {
	role := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cm-editor",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "watch", "list", "create", "update"},
				Resources: []string{"configmaps"},
				APIGroups: []string{""},
			},
		},
	}

	roleBinding := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cm-editor",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: "default",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     "cm-editor",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	if _, err := rbacv1Client.Roles(ns).Create(ctx, &role, metav1.CreateOptions{}); err != nil {
		return err
	}

	if _, err := rbacv1Client.RoleBindings(ns).Create(ctx, &roleBinding, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}
