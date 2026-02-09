package domain


func (oi *OrderItem) CanBeCancelled() bool {
    if oi.Status == OrderItemStatusCancelled || oi.Status == OrderItemStatusRefunded {
        return false
    }
    // Ensure order-level checks are applied before calling
    return true
}
