import os, psycopg2
from psycopg2.extras import RealDictCursor
from datetime import datetime

def get_db_connection():
    return psycopg2.connect(
        host=os.getenv("DB_HOST", "db_bonus"),
        database=os.getenv("DB_NAME", "bonus"),
        user=os.getenv("DB_USER", "program"),
        password=os.getenv("DB_PASSWORD", "test"),
        port=os.getenv("DB_PORT", "5432")
    )

def get_privilege_with_history(username: str):
    conn = get_db_connection()
    cur = conn.cursor(cursor_factory=RealDictCursor)
    
    # Получаем саму привилегию
    cur.execute("SELECT id, balance, status FROM privilege WHERE username = %s", (username,))
    privilege = cur.fetchone()
    
    if not privilege:
        cur.close()
        conn.close()
        return None

    # Получаем историю
    cur.execute("""
        SELECT datetime as "date", ticket_uid as "ticketUid", 
               balance_diff as "balanceDiff", operation_type as "operationType"
        FROM privilege_history WHERE privilege_id = %s
    """, (privilege['id'],))
    history = cur.fetchall()
    
    cur.close()
    conn.close()
    return {"balance": privilege['balance'], "status": privilege['status'], "history": history}

def process_bonus_operation(username: str, ticket_uid: str, price: int, paid_from_balance: bool):
    conn = get_db_connection()
    cur = conn.cursor(cursor_factory=RealDictCursor)
    
    # 1. Получаем текущий баланс
    cur.execute("SELECT id, balance, status FROM privilege WHERE username = %s", (username,))
    priv = cur.fetchone()
    
    # ЗАЩИТА: Если пользователя нет в таблице privilege
    if priv is None:
        # Вариант А: Автоматическое создание записи (рекомендуется для тестов)
        cur.execute(
            "INSERT INTO privilege (username, status, balance) VALUES (%s, 'BRONZE', 0) RETURNING id, balance, status",
            (username,)
        )
        priv = cur.fetchone()
        # Вариант Б: Если создание не предусмотрено, можно бросить исключение:
        # raise HTTPException(status_code=404, detail="User privilege profile not found")

    paid_by_bonuses = 0
    balance_diff = 0
    op_type = ""

    if paid_from_balance:
        # Списание: используем доступные бонусы
        paid_by_bonuses = min(priv['balance'], price)
        balance_diff = -paid_by_bonuses
        op_type = 'DEBIT_THE_ACCOUNT'
    else:
        # Начисление: 10% от цены
        balance_diff = int(price * 0.1)
        paid_by_bonuses = 0
        op_type = 'FILL_IN_BALANCE'

    # 2. Обновляем баланс
    cur.execute(
        "UPDATE privilege SET balance = balance + %s WHERE id = %s RETURNING balance, status",
        (balance_diff, priv['id'])
    )
    updated_priv = cur.fetchone()

    # 3. Записываем историю
    cur.execute("""
        INSERT INTO privilege_history (privilege_id, ticket_uid, datetime, balance_diff, operation_type)
        VALUES (%s, %s, %s, %s, %s)
    """, (priv['id'], ticket_uid, datetime.now(), abs(balance_diff), op_type))

    conn.commit()
    cur.close()
    conn.close()
    
    return {
        "paidByBonuses": paid_by_bonuses,
        "balanceDiff": balance_diff,
        "privilege": {
            "balance": updated_priv['balance'], 
            "status": updated_priv['status']
        }
    }