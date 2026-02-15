package domain


func (oi *OrderItem) CanBeCancelled() bool {
    if oi.Status == OrderItemStatusCancelled || oi.Status == OrderItemStatusRefunded {
        return false
    }
    return true
}
