import { PromotionApplication } from '@/components/courier/growth/PromotionApplication';
import { CourierPermissionGuard } from '@/components/courier/CourierPermissionGuard';

export default function PromotionApplyPage() {
  return (
    <CourierPermissionGuard requiredLevel={1}>
      <div className="container mx-auto py-6">
        <PromotionApplication />
      </div>
    </CourierPermissionGuard>
  );
}