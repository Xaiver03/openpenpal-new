import { PromotionManagement } from '@/components/courier/growth/PromotionManagement';
import { CourierPermissionGuard } from '@/components/courier/CourierPermissionGuard';

export default function PromotionManagePage() {
  return (
    <CourierPermissionGuard requiredLevel={3}>
      <div className="container mx-auto py-6">
        <PromotionManagement />
      </div>
    </CourierPermissionGuard>
  );
}