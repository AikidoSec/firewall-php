import requests
import time
import sys
from testlib import *

def expect_blocked(action, tenant_id="org_123"):
    data = {"action": action, "tenantId": tenant_id}
    response = php_server_post("/", data)
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Zen IDOR protection")
    return response

def expect_allowed(action, tenant_id="org_123"):
    data = {"action": action, "tenantId": tenant_id}
    response = php_server_post("/", data)
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "OK")
    return response

def run_test():
    # SELECT without tenant filter should be blocked
    expect_blocked("select_without_filter")

    # SELECT with correct tenant filter should be allowed
    expect_allowed("select_with_correct_filter")

    # SELECT with wrong tenant ID value should be blocked
    expect_blocked("select_with_wrong_filter")

    # SELECT on excluded table should be allowed
    expect_allowed("select_excluded_table")

    # INSERT without tenant column should be blocked
    expect_blocked("insert_without_tenant_column")

    # INSERT with correct tenant value should be allowed
    expect_allowed("insert_with_correct_tenant")

    # INSERT with wrong tenant value should be blocked
    expect_blocked("insert_with_wrong_tenant")

    # UPDATE without tenant filter should be blocked
    expect_blocked("update_without_filter")

    # UPDATE with correct tenant filter should be allowed
    expect_allowed("update_with_correct_filter")

    # DELETE without tenant filter should be blocked
    expect_blocked("delete_without_filter")

    # DELETE with correct tenant filter should be allowed
    expect_allowed("delete_with_correct_filter")

    # without_idor_protection callback should bypass check
    expect_allowed("without_idor_protection")

    # Missing set_tenant_id should be blocked
    response = php_server_post("/", {"action": "select_with_correct_filter"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "setTenantId() was not called")

    # DDL statements should be allowed
    expect_allowed("ddl_statement")

    # Transaction statements should be allowed
    expect_allowed("transaction")

    # Prepared statement with correct positional param should be allowed
    expect_allowed("prepared_positional_correct")

    # Prepared statement with wrong positional param should be blocked
    expect_blocked("prepared_positional_wrong")

    # Prepared statement with correct named param should be allowed
    expect_allowed("prepared_named_correct")

    # Prepared statement with wrong named param should be blocked
    expect_blocked("prepared_named_wrong")

    # Prepared INSERT with correct param should be allowed
    expect_allowed("prepared_insert_correct")

    # Prepared INSERT with wrong param should be blocked
    expect_blocked("prepared_insert_wrong")

    # bindValue with correct value should be allowed
    expect_allowed("bind_value_correct")

    # bindValue with wrong value should be blocked
    expect_blocked("bind_value_wrong")

    # bindParam with correct value should be allowed
    expect_allowed("bind_param_correct")

    # bindParam with wrong value should be blocked
    expect_blocked("bind_param_wrong")

    # bindValue positional with correct value should be allowed
    expect_allowed("bind_value_positional_correct")

    # bindValue positional with wrong value should be blocked
    expect_blocked("bind_value_positional_wrong")

    # IDOR violations should throw even in detect-only mode (block=false)
    apply_config("change_config_disable_blocking.json")
    expect_blocked("select_without_filter")
    expect_blocked("insert_without_tenant_column")

if __name__ == "__main__":
    load_test_args()
    run_test()
