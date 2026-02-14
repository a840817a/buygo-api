#!/bin/bash
set -e

echo "🔄 Starting Project → GroupBuy refactoring..."

# ============================================
# Phase 1: Rename Domain Layer
# ============================================
echo "📁 Phase 1: Renaming domain/project → domain/groupbuy"
if [ -d "internal/domain/project" ]; then
    mv internal/domain/project internal/domain/groupbuy
    echo "   ✓ Directory renamed"
fi

# ============================================
# Phase 2: Rename Files
# ============================================
echo "📄 Phase 2: Renaming files"

# Service layer
if [ -f "internal/service/project_service.go" ]; then
    mv internal/service/project_service.go internal/service/groupbuy_service.go
fi
if [ -f "internal/service/project_service_test.go" ]; then
    mv internal/service/project_service_test.go internal/service/groupbuy_service_test.go
fi
if [ -f "internal/service/project_service_template.go" ]; then
    mv internal/service/project_service_template.go internal/service/groupbuy_service_template.go
fi
if [ -f "internal/service/project_security_test.go" ]; then
    mv internal/service/project_security_test.go internal/service/groupbuy_security_test.go
fi
if [ -f "internal/service/project_business_test.go" ]; then
    mv internal/service/project_business_test.go internal/service/groupbuy_business_test.go
fi

# Repository layer
if [ -f "internal/adapter/repository/memory/project_repo.go" ]; then
    mv internal/adapter/repository/memory/project_repo.go internal/adapter/repository/memory/groupbuy_repo.go
fi
if [ -f "internal/adapter/repository/postgres/project_repo.go" ]; then
    mv internal/adapter/repository/postgres/project_repo.go internal/adapter/repository/postgres/groupbuy_repo.go
fi
if [ -f "internal/adapter/repository/postgres/project_repo_test.go" ]; then
    mv internal/adapter/repository/postgres/project_repo_test.go internal/adapter/repository/postgres/groupbuy_repo_test.go
fi
if [ -f "internal/adapter/repository/postgres/model/project.go" ]; then
    mv internal/adapter/repository/postgres/model/project.go internal/adapter/repository/postgres/model/groupbuy.go
fi

# Handler layer
if [ -f "internal/adapter/handler/project_handler.go" ]; then
    mv internal/adapter/handler/project_handler.go internal/adapter/handler/groupbuy_handler.go
fi

echo "   ✓ Files renamed"

# ============================================
# Phase 3: Content Replacement
# ============================================
echo "🔍 Phase 3: Replacing content in all Go files"

# Find all Go files and replace content
find . -name "*.go" -type f -not -path "./api/*" -not -path "./.git/*" | while read file; do
    # Domain/package references
    sed -i.bak \
        -e 's|github.com/buygo/buygo-api/internal/domain/project|github.com/buygo/buygo-api/internal/domain/groupbuy|g' \
        -e 's|"project"|"groupbuy"|g' \
        "$file"

    # Type names
    sed -i.bak \
        -e 's/\bProjectRepository\b/GroupBuyRepository/g' \
        -e 's/\bProjectService\b/GroupBuyService/g' \
        -e 's/\bProjectHandler\b/GroupBuyHandler/g' \
        -e 's/\bNewProjectService\b/NewGroupBuyService/g' \
        -e 's/\bNewProjectRepository\b/NewGroupBuyRepository/g' \
        -e 's/\bNewProjectHandler\b/NewGroupBuyHandler/g' \
        "$file"

    # Function/method names containing Project (be careful with Product)
    sed -i.bak \
        -e 's/\bCreateProject\b/CreateGroupBuy/g' \
        -e 's/\bGetProject\b/GetGroupBuy/g' \
        -e 's/\bListProjects\b/ListGroupBuys/g' \
        -e 's/\bUpdateProject\b/UpdateGroupBuy/g' \
        -e 's/\bListManagerProjects\b/ListManagerGroupBuys/g' \
        -e 's/\bGetMyProjectOrder\b/GetMyGroupBuyOrder/g' \
        -e 's/\bListProjectOrders\b/ListGroupBuyOrders/g' \
        "$file"

    # Variable names (common patterns)
    sed -i.bak \
        -e 's/\bprojectRepo\b/groupBuyRepo/g' \
        -e 's/\bprojectService\b/groupBuyService/g' \
        -e 's/\bprojectHandler\b/groupBuyHandler/g' \
        "$file"

    # Comments and strings (case-sensitive)
    sed -i.bak \
        -e 's/Project repository/GroupBuy repository/g' \
        -e 's/Project service/GroupBuy service/g' \
        -e 's/project service/group buy service/g' \
        -e 's/project repository/group buy repository/g' \
        "$file"

    # Database table references
    sed -i.bak \
        -e 's/"projects"/"group_buys"/g' \
        -e 's/`projects`/`group_buys`/g' \
        "$file"

    # Struct fields in models
    sed -i.bak \
        -e 's/\bProjectID\b/GroupBuyID/g' \
        -e 's/project_id/group_buy_id/g' \
        "$file"

    rm -f "${file}.bak"
done

echo "   ✓ Content replaced in Go files"

# ============================================
# Phase 4: Update Proto-generated files
# ============================================
echo "🔄 Phase 4: Removing old proto-generated files"
rm -f api/v1/project.pb.go
rm -f api/v1/buygov1connect/project.connect.go

echo "   ✓ Old proto files removed (will regenerate later)"

# ============================================
# Phase 5: Clean up backup files
# ============================================
echo "🧹 Phase 5: Cleaning up backup files"
find . -name "*.bak" -type f -delete
echo "   ✓ Backup files cleaned"

echo ""
echo "✅ Backend refactoring complete!"
echo ""
echo "📝 Next steps:"
echo "  1. Install buf: go install github.com/bufbuild/buf/cmd/buf@latest"
echo "  2. Regenerate proto: buf generate"
echo "  3. Run tests: go test ./..."
