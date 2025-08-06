import { CourierGrowthPath } from '@/components/courier/growth/CourierGrowthPath';
import { CourierPermissionGuard } from '@/components/courier/CourierPermissionGuard';

export default function CourierGrowthPage() {
  return (
    <CourierPermissionGuard requiredLevel={1}>
      <div className="container mx-auto py-6">
        <h1 className="text-2xl font-bold mb-6">我的成长路径</h1>
        <CourierGrowthPath />
      </div>
    </CourierPermissionGuard>
  );
}