export interface ReportRange {
  date_from: string
  date_to: string
  timezone: 'America/La_Paz'
}

export interface OperationalSummary {
  orders_received: number
  orders_delivered: number
  orders_cancelled: number
  active_orders: number
  ready_for_pickup: number
  overdue_orders: number
  average_cycle_minutes: number
}

export interface StatusMetric { status: string, count: number }
export interface DailyOperationalMetric { date: string, received: number, delivered: number }
export interface ServiceMetric { service_name: string, uses: number }

export interface OperationalReport {
  range: ReportRange
  summary: OperationalSummary
  status: StatusMetric[]
  daily: DailyOperationalMetric[]
  top_services: ServiceMetric[]
}

export interface FinancialSummary {
  revenue: string
  payments: number
  average_ticket: string
  voided_amount: string
  voided_count: number
}

export interface PaymentMethodMetric { method: string, count: number, amount: string }
export interface DailyFinancialMetric { date: string, payments: number, amount: string }

export interface FinancialReport {
  range: ReportRange
  summary: FinancialSummary
  methods: PaymentMethodMetric[]
  daily: DailyFinancialMetric[]
}

export interface DashboardReport {
  operational: OperationalReport
  financial: FinancialReport
}
