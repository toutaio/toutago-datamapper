# Custom Actions Example

This example demonstrates **custom actions** beyond standard CRUD operations with toutago-datamapper.

## What This Example Shows

- **List Actions**: Query with filters
- **Aggregations**: Calculate summaries (SUM, COUNT, etc.)
- **Stored Procedures**: Execute complex database procedures
- **Custom Operations**: Domain-specific business logic
- **Parameterized Actions**: Dynamic queries with parameters

## Action Types

### 1. **Query Actions**
Retrieve data with custom filtering, sorting, and limiting.

```yaml
actions:
  list-all:
    statement: "transactions/*.json"
    result:
      type: Transaction
      multi: true
```

### 2. **Filtered Actions**
Query with parameters and filters.

```yaml
actions:
  by-account:
    statement: "transactions/*.json"
    parameters:
      - name: account_id
        type: string
    filters:
      - field: account_id
        operator: equals
        parameter: account_id
```

### 3. **Aggregation Actions**
Calculate statistics and summaries.

```yaml
actions:
  account-summary:
    aggregations:
      - function: sum
        field: amount
        alias: total_credits
      - function: count
        field: id
        alias: transaction_count
```

### 4. **Stored Procedures**
Execute database stored procedures.

```yaml
actions:
  reconcile-balance:
    statement: "CALL reconcile_account_balance({account_id})"
    type: procedure
    parameters:
      - name: account_id
        type: string
```

## Running the Example

```bash
cd examples/custom-actions
go run main.go
```

## Expected Output

```
=== Custom Actions Example ===

1. Creating sample transactions...
   ✓ Created 8 transactions

2. Listing all transactions...
   ✓ Found 8 transactions

3. Getting transactions for account acc-100...
   ✓ Found 5 transactions for acc-100
      1. credit - Initial deposit $1000.00 (Balance: $1000.00)
      2. debit - ATM withdrawal $50.00 (Balance: $950.00)
      3. credit - Salary payment $200.00 (Balance: $1150.00)
      4. debit - Grocery shopping $75.50 (Balance: $1074.50)
      5. debit - Restaurant $30.00 (Balance: $1044.50)

4. Calculating account summary for acc-100...
   Account Summary:
   • Total Credits: $1200.00
   • Total Debits: $155.50
   • Transaction Count: 5
   • Current Balance: $1044.50

5. Getting 3 most recent transactions...
   ✓ Recent transactions:
      1. [acc-101] Freelance payment - $250.00
      2. [acc-101] Online purchase - $100.00
      3. [acc-101] Initial deposit - $500.00

6. Getting all credit transactions...
   ✓ Found 4 credit transactions
   ✓ Total credit amount: $1950.00

7. Executing balance reconciliation (stored procedure simulation)...
   ✓ Balance reconciled:
      Expected: $1044.50
      Actual: $1044.50
      Status: OK

=== Custom Actions Demonstrated ===
```

## Usage in Code

### Simple List Action
```go
result, err := mapper.Execute(ctx, "transactions.list-all", nil)
transactions := result.([]interface{})
```

### Parameterized Action
```go
result, err := mapper.Execute(ctx, "transactions.by-account", map[string]interface{}{
    "account_id": "acc-100",
})
```

### Aggregation Action
```go
result, err := mapper.Execute(ctx, "transactions.account-summary", map[string]interface{}{
    "account_id": "acc-100",
})
summary := result.(map[string]interface{})
totalCredits := summary["total_credits"].(float64)
```

### Stored Procedure
```go
result, err := mapper.Execute(ctx, "transactions.reconcile-balance", map[string]interface{}{
    "account_id": "acc-100",
})
```

## Real-World Use Cases

### 1. **Reporting Queries**
```yaml
actions:
  monthly-sales-report:
    statement: "SELECT * FROM sales WHERE month = {month} AND year = {year}"
    parameters:
      - name: month
        type: int
      - name: year
        type: int
    aggregations:
      - function: sum
        field: total
        alias: monthly_revenue
      - function: avg
        field: total
        alias: average_order_value
```

### 2. **Search Actions**
```yaml
actions:
  search-products:
    statement: "SELECT * FROM products WHERE name LIKE {query}"
    parameters:
      - name: query
        type: string
    filters:
      - field: active
        value: true
    sort:
      - field: name
        direction: asc
    limit: 50
```

### 3. **Complex Business Logic**
```yaml
actions:
  close-month:
    statement: "CALL close_accounting_month({year}, {month})"
    type: procedure
    parameters:
      - name: year
        type: int
      - name: month
        type: int
    result:
      type: map
      properties:
        - field: status
        - field: records_processed
        - field: errors
```

### 4. **Batch Operations**
```yaml
actions:
  process-pending-orders:
    statement: "CALL process_orders()"
    type: procedure
    result:
      type: map
      properties:
        - field: processed_count
        - field: failed_count
        - field: execution_time
```

## Advanced Features

### Filtering
```yaml
filters:
  - field: status
    operator: equals
    parameter: status
  - field: created_at
    operator: greater_than
    parameter: start_date
  - field: amount
    operator: between
    parameters: [min_amount, max_amount]
```

### Sorting
```yaml
sort:
  - field: created_at
    direction: desc
  - field: amount
    direction: asc
```

### Pagination
```yaml
limit: 100
offset: "{page_offset}"
```

### Aggregations
```yaml
aggregations:
  - function: sum
    field: amount
    alias: total
  - function: avg
    field: amount
    alias: average
  - function: min
    field: amount
    alias: minimum
  - function: max
    field: amount
    alias: maximum
  - function: count
    field: id
    alias: count
```

## Database-Specific Examples

### MySQL Stored Procedure
```yaml
actions:
  calculate-commissions:
    statement: "CALL calculate_sales_commissions({period})"
    type: procedure
    parameters:
      - name: period
        type: string
```

### PostgreSQL Function
```yaml
actions:
  generate-invoice:
    statement: "SELECT * FROM generate_invoice({order_id})"
    type: function
    parameters:
      - name: order_id
        type: int
```

### MongoDB Aggregation Pipeline
```yaml
actions:
  top-customers:
    statement: |
      {
        $group: {
          _id: "$customer_id",
          total_spent: { $sum: "$amount" },
          order_count: { $sum: 1 }
        },
        $sort: { total_spent: -1 },
        $limit: 10
      }
    type: aggregation
```

## Testing Custom Actions

```go
func TestCustomAction(t *testing.T) {
    mapper, _ := engine.NewMapper("config.yaml")
    
    // Execute action
    result, err := mapper.Execute(ctx, "transactions.by-account", map[string]interface{}{
        "account_id": "test-123",
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    
    txns := result.([]interface{})
    assert.Greater(t, len(txns), 0)
}
```

## Performance Considerations

1. **Use Indexes**: Ensure filtered fields are indexed
2. **Limit Results**: Always use LIMIT for large datasets
3. **Cache Results**: Cache frequently-accessed aggregations
4. **Optimize Queries**: Use EXPLAIN to analyze query plans
5. **Batch Processing**: Process large datasets in chunks

## Benefits

### Encapsulation
- Complex queries hidden behind simple actions
- Business logic in configuration, not code
- Easy to modify without code changes

### Reusability
- Actions can be called from multiple places
- Consistent query logic across application
- Reduced code duplication

### Testing
- Actions can be tested independently
- Mock different data sources easily
- Clear separation of concerns

### Maintainability
- Centralized query definitions
- Easy to update for performance
- Version control for queries

## Best Practices

1. **Name Clearly**: Use descriptive action names
2. **Document**: Add descriptions to all actions
3. **Validate**: Check parameters before execution
4. **Handle Errors**: Provide meaningful error messages
5. **Optimize**: Monitor and optimize slow actions
6. **Version**: Track action changes over time
7. **Test**: Write tests for all custom actions

## Next Steps

- Implement with real databases (MySQL, PostgreSQL)
- Add caching for expensive aggregations
- Create action composition (one action calls another)
- Build a query builder UI
- Add action monitoring and metrics
