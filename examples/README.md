# toutago-datamapper Examples

This directory contains comprehensive examples demonstrating all major features of toutago-datamapper.

## 📚 Examples Overview

### 1. [Simple CRUD](simple-crud/) - Start Here!
**Difficulty:** Beginner  
**Time:** 5 minutes

Learn the basics of toutago-datamapper with a straightforward example:
- Creating a mapper from configuration
- Registering adapters
- Performing CRUD operations (Create, Read, Update, Delete)
- Working with domain objects

**What you'll learn:**
- Basic mapper usage
- Configuration structure
- Filesystem adapter
- Type-safe operations

**Best for:**
- New users
- Quick start
- Understanding core concepts

---

### 2. [Multi-Source CQRS](multi-source/) - Architecture Patterns
**Difficulty:** Intermediate  
**Time:** 10 minutes

Implement CQRS (Command Query Responsibility Segregation) pattern:
- Multiple data sources (primary + cache)
- Read/write separation
- Cache invalidation strategies
- Performance optimization

**What you'll learn:**
- CQRS architecture
- Multi-source configuration
- Cache strategies
- Performance patterns

**Best for:**
- Scalable applications
- High-performance systems
- Read-heavy workloads
- Distributed systems

---

### 3. [Bulk Operations](bulk-operations/) - High Performance
**Difficulty:** Intermediate  
**Time:** 10 minutes

Process large datasets efficiently:
- Bulk insert (100+ records at once)
- Bulk update and delete
- Performance measurement
- Memory-efficient patterns

**What you'll learn:**
- Bulk operation configuration
- Performance optimization
- Batch processing
- Resource management

**Best for:**
- Data imports/exports
- Batch processing
- ETL pipelines
- Large datasets

---

### 4. [Credentials Management](credentials/) - Security Best Practices
**Difficulty:** Intermediate  
**Time:** 15 minutes

Secure credential handling:
- Environment variable resolution
- Separate secrets from configuration
- Multi-environment setup
- Git-safe practices

**What you'll learn:**
- Credential resolution
- Environment variables
- Security best practices
- Multi-environment deployment

**Best for:**
- Production deployments
- Multi-environment setups
- Security-conscious projects
- Team collaboration

---

### 5. [Custom Actions](custom-actions/) - Advanced Features
**Difficulty:** Advanced  
**Time:** 15 minutes

Beyond CRUD operations:
- Custom query actions
- Aggregations (SUM, COUNT, AVG)
- Stored procedures
- Complex business logic

**What you'll learn:**
- Custom action configuration
- Aggregation queries
- Stored procedures
- Advanced patterns

**Best for:**
- Complex queries
- Business logic encapsulation
- Reporting systems
- Advanced use cases

---

## 🚀 Quick Start

### Run Any Example

```bash
# Clone the repository
git clone https://github.com/toutaio/toutago-datamapper.git
cd toutago-datamapper/examples

# Run simple CRUD
cd simple-crud
go run main.go

# Run multi-source CQRS
cd ../multi-source
go run main.go

# Run bulk operations
cd ../bulk-operations
go run main.go

# Run credentials (set env vars first)
cd ../credentials
source .env
go run main.go

# Run custom actions
cd ../custom-actions
go run main.go
```

---

## 📖 Learning Path

### For Beginners
1. **Start:** [Simple CRUD](simple-crud/) - Learn the basics
2. **Next:** [Credentials](credentials/) - Secure your app
3. **Then:** [Bulk Operations](bulk-operations/) - Scale up

### For Intermediate Developers
1. **Start:** [Multi-Source](multi-source/) - Architecture patterns
2. **Next:** [Custom Actions](custom-actions/) - Advanced features
3. **Then:** [Bulk Operations](bulk-operations/) - Performance

### For Advanced Users
1. **Start:** [Custom Actions](custom-actions/) - Complex operations
2. **Next:** [Multi-Source](multi-source/) - CQRS patterns
3. **Then:** Create your own adapter!

---

## 🎯 Use Case Guide

### Building a REST API?
→ Start with [Simple CRUD](simple-crud/)  
→ Add [Credentials](credentials/) for security  
→ Use [Custom Actions](custom-actions/) for complex queries

### High-Performance System?
→ Start with [Multi-Source](multi-source/) for caching  
→ Add [Bulk Operations](bulk-operations/) for batch processing  
→ Optimize with read replicas

### Data Migration?
→ Start with [Bulk Operations](bulk-operations/)  
→ Add [Credentials](credentials/) for different environments  
→ Use batch processing patterns

### Microservices?
→ Start with [Simple CRUD](simple-crud/) per service  
→ Add [Multi-Source](multi-source/) for CQRS  
→ Use [Credentials](credentials/) for env-specific config

---

## 💡 Key Concepts Demonstrated

### Configuration-Driven
All examples show how to define mappings in YAML:
```yaml
namespace: myapp
sources:
  database:
    adapter: filesystem
    connection: "./data"
mappings:
  user-crud:
    object: User
    source: database
    operations:
      fetch:
        statement: "users/{id}.json"
```

### Zero Dependencies
Domain objects have no library dependencies:
```go
type User struct {
    ID    string
    Name  string
    Email string
}
```

### Type Safety
All operations are type-safe:
```go
var user User
mapper.Fetch(ctx, "users.user-crud", params, &user)
```

### Pluggable Adapters
Easy to swap data sources:
```yaml
# Development
sources:
  db:
    adapter: filesystem

# Production
sources:
  db:
    adapter: postgres
```

---

## 🔧 Example Structure

Each example contains:

```
example-name/
├── main.go          # Runnable code
├── config.yaml      # Configuration file
├── README.md        # Detailed documentation
└── data/            # Generated by running example
```

### main.go
- Demonstrates feature usage
- Includes comments
- Shows error handling
- Prints clear output

### config.yaml
- Complete configuration
- Commented sections
- Ready to use
- Easy to modify

### README.md
- What the example shows
- How to run it
- Expected output
- Real-world use cases
- Best practices
- Next steps

---

## 🌟 Features Demonstrated

### Core Features
- ✅ CRUD operations (Create, Read, Update, Delete)
- ✅ Bulk operations (Insert, Update, Delete many)
- ✅ Custom actions (Queries, Aggregations, Procedures)
- ✅ Multi-source support (CQRS, caching)
- ✅ Credential management (Environment variables)

### Advanced Features
- ✅ CQRS pattern implementation
- ✅ Cache invalidation strategies
- ✅ Stored procedure simulation
- ✅ Aggregation queries
- ✅ Filtered queries
- ✅ Parameterized actions
- ✅ Batch processing

### Production Features
- ✅ Environment variable resolution
- ✅ Multi-environment setup
- ✅ Security best practices
- ✅ Error handling
- ✅ Performance monitoring
- ✅ Resource management

---

## 📊 Comparison Matrix

| Feature | Simple CRUD | Multi-Source | Bulk Ops | Credentials | Custom Actions |
|---------|-------------|--------------|----------|-------------|----------------|
| Difficulty | ⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| Time | 5 min | 10 min | 10 min | 15 min | 15 min |
| CRUD | ✅ | ✅ | ✅ | ✅ | ✅ |
| CQRS | ❌ | ✅ | ❌ | ❌ | ❌ |
| Bulk | ❌ | ❌ | ✅ | ❌ | ❌ |
| Security | ❌ | ❌ | ❌ | ✅ | ❌ |
| Actions | ❌ | ❌ | ❌ | ❌ | ✅ |
| Best For | Learning | Architecture | Performance | Security | Advanced |

---

## 🔗 Related Documentation

- [Main README](../README.md) - Project overview
- [Implementation Plan](../IMPLEMENTATION_PLAN.md) - Roadmap
- [Quick Start](../QUICKSTART.md) - Getting started

---

## 🤝 Contributing Examples

Want to add an example? Great! Examples should:

1. **Be Self-Contained**: Run without external dependencies
2. **Be Well-Documented**: Include comprehensive README
3. **Show Real Use Cases**: Solve actual problems
4. **Follow Patterns**: Match existing example structure
5. **Include Output**: Show what users should expect

### Example Ideas
- MongoDB adapter example
- Redis cache example
- PostgreSQL with transactions
- Event sourcing pattern
- Real-time sync example
- GraphQL integration
- REST API complete example

---

## 📝 Example Code Standards

All examples follow these standards:

### Code Quality
- ✅ Clear variable names
- ✅ Comprehensive comments
- ✅ Error handling shown
- ✅ Output formatting
- ✅ Context usage

### Configuration
- ✅ Well-organized YAML
- ✅ Inline comments
- ✅ Clear naming
- ✅ Complete mappings

### Documentation
- ✅ Clear purpose statement
- ✅ Step-by-step instructions
- ✅ Expected output shown
- ✅ Real-world use cases
- ✅ Next steps provided

---

## 🎓 Learning Resources

### After Running Examples
1. Read the [Implementation Plan](../IMPLEMENTATION_PLAN.md)
2. Review the main [README](../README.md) for configuration options
3. Check the [Quick Start Guide](../QUICKSTART.md)

### Building Your Own
1. Start with simple-crud template
2. Modify configuration for your needs
3. Add your domain objects
4. Implement your adapter (or use existing)
5. Test thoroughly

---

## ❓ FAQ

**Q: Which example should I start with?**  
A: Start with [Simple CRUD](simple-crud/) - it covers all basics.

**Q: Can I use these in production?**  
A: Examples use filesystem adapter. For production, use MySQL/PostgreSQL adapters.

**Q: How do I create my own adapter?**  
A: Implement the `adapter.Adapter` interface. See [filesystem adapter](../filesystem/) as reference.

**Q: Are there database examples?**  
A: MySQL and PostgreSQL adapters coming in Phase 6. Examples will follow.

**Q: Can I mix multiple patterns?**  
A: Yes! Combine CQRS + Bulk + Custom Actions as needed.

**Q: How do I handle migrations?**  
A: Use bulk operations to migrate data between sources.

---

## 🚀 Next Steps

1. **Run all examples** to see different patterns
2. **Modify configurations** to match your needs
3. **Implement your domain objects**
4. **Choose/create your adapter**
5. **Build your application!**

---

## 📞 Getting Help

- **Documentation**: See [docs](../docs/) and main [README](../README.md)
- **Issues**: [GitHub Issues](https://github.com/toutaio/toutago-datamapper/issues)
- **Examples Not Working?**: Check Go version (requires 1.21+)

---

**Happy Coding! 🎉**

All examples are production-quality and ready to use as templates for your projects.
