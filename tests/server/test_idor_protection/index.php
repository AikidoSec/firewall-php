<?php

\aikido\enable_idor_protection("tenant_id", ["users"]);

$requestBody = file_get_contents('php://input');
$data = json_decode($requestBody, true);

if (isset($data['tenantId'])) {
    \aikido\set_tenant_id($data['tenantId']);
}

try {
    $pdo = new PDO("sqlite::memory:");
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    $pdo->exec("CREATE TABLE IF NOT EXISTS orders (id INTEGER PRIMARY KEY, name TEXT, tenant_id TEXT)");
    $pdo->exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT, tenant_id TEXT)");

    $action = $data['action'] ?? '';

    switch ($action) {
        case 'select_without_filter':
            $pdo->query("SELECT * FROM orders");
            break;

        case 'select_with_correct_filter':
            $pdo->query("SELECT * FROM orders WHERE tenant_id = '" . $data['tenantId'] . "'");
            break;

        case 'select_with_wrong_filter':
            $pdo->query("SELECT * FROM orders WHERE tenant_id = 'wrong_tenant'");
            break;

        case 'select_excluded_table':
            $pdo->query("SELECT * FROM users");
            break;

        case 'insert_without_tenant_column':
            $pdo->exec("INSERT INTO orders (name) VALUES ('Widget')");
            break;

        case 'insert_with_correct_tenant':
            $pdo->exec("INSERT INTO orders (name, tenant_id) VALUES ('Widget', '" . $data['tenantId'] . "')");
            break;

        case 'insert_with_wrong_tenant':
            $pdo->exec("INSERT INTO orders (name, tenant_id) VALUES ('Widget', 'wrong_tenant')");
            break;

        case 'update_without_filter':
            $pdo->exec("UPDATE orders SET name = 'Updated'");
            break;

        case 'update_with_correct_filter':
            $pdo->exec("UPDATE orders SET name = 'Updated' WHERE tenant_id = '" . $data['tenantId'] . "'");
            break;

        case 'delete_without_filter':
            $pdo->exec("DELETE FROM orders");
            break;

        case 'delete_with_correct_filter':
            $pdo->exec("DELETE FROM orders WHERE tenant_id = '" . $data['tenantId'] . "'");
            break;

        case 'without_idor_protection':
            \aikido\without_idor_protection(function() use ($pdo) {
                $pdo->query("SELECT * FROM orders");
            });
            break;

        case 'ddl_statement':
            $pdo->exec("CREATE TABLE IF NOT EXISTS temp_table (id INTEGER)");
            break;

        case 'transaction':
            $pdo->exec("BEGIN");
            $pdo->exec("COMMIT");
            break;

        case 'prepared_positional_correct':
            $stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = ?");
            $stmt->execute([$data['tenantId']]);
            break;

        case 'prepared_positional_wrong':
            $stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = ?");
            $stmt->execute(['wrong_tenant']);
            break;

        case 'prepared_named_correct':
            $stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = :tid");
            $stmt->execute([':tid' => $data['tenantId']]);
            break;

        case 'prepared_named_wrong':
            $stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = :tid");
            $stmt->execute([':tid' => 'wrong_tenant']);
            break;

        case 'prepared_insert_correct':
            $stmt = $pdo->prepare("INSERT INTO orders (name, tenant_id) VALUES (?, ?)");
            $stmt->execute(['Widget', $data['tenantId']]);
            break;

        case 'prepared_insert_wrong':
            $stmt = $pdo->prepare("INSERT INTO orders (name, tenant_id) VALUES (?, ?)");
            $stmt->execute(['Widget', 'wrong_tenant']);
            break;

        case 'bind_value_correct':
            $stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = :tid");
            $stmt->bindValue(':tid', $data['tenantId']);
            $stmt->execute();
            break;

        case 'bind_value_wrong':
            $stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = :tid");
            $stmt->bindValue(':tid', 'wrong_tenant');
            $stmt->execute();
            break;

        case 'bind_param_correct':
            $stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = :tid");
            $tid = $data['tenantId'];
            $stmt->bindParam(':tid', $tid);
            $stmt->execute();
            break;

        case 'bind_param_wrong':
            $stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = :tid");
            $tid = 'wrong_tenant';
            $stmt->bindParam(':tid', $tid);
            $stmt->execute();
            break;

        case 'bind_value_positional_correct':
            $stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = ?");
            $stmt->bindValue(1, $data['tenantId']);
            $stmt->execute();
            break;

        case 'bind_value_positional_wrong':
            $stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = ?");
            $stmt->bindValue(1, 'wrong_tenant');
            $stmt->execute();
            break;

        default:
            echo "Unknown action";
            break;
    }

    echo "OK";
} catch (\Exception $e) {
    http_response_code(500);
    echo $e->getMessage();
}

?>
