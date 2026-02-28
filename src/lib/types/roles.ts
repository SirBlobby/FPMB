export enum RoleFlag {
	Viewer = 1 << 0, // 1
	Editor = 1 << 1, // 2
	Admin  = 1 << 2, // 4
	Owner  = 1 << 3, // 8
}

export function hasPermission(userRole: number, requiredRole: number): boolean {
	return (userRole & requiredRole) === requiredRole;
}
