import os
from fastapi import FastAPI, Header, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
from dotenv import load_dotenv

from .clients import FlightClient, TicketClient, BonusClient

load_dotenv()

app = FastAPI(title="Gateway Service")

# Настройка CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Инициализация клиентов
flight_client = FlightClient(os.getenv("FLIGHT_SERVICE_HOST"))
ticket_client = TicketClient(os.getenv("TICKET_SERVICE_HOST"))
bonus_client = BonusClient(os.getenv("BONUS_SERVICE_HOST"))

@app.get("/api/v1/flights")
async def get_flights(page: int = 0, size: int = 10):
    resp = await flight_client.get_flights(page, size)
    return resp.json()

@app.get("/api/v1/tickets")
async def get_user_tickets(x_user_name: str = Header(...)):
    # 1. Получаем список билетов
    t_resp = await ticket_client.get_tickets(x_user_name)
    if t_resp.status_code != 200:
        return []
    
    tickets = t_resp.json()
    result = []

    # 2. Обогащаем данными о рейсах
    for t in tickets:
        f_resp = await flight_client.get_flight(t["flightNumber"])
        f_data = f_resp.json() if f_resp.status_code == 200 else {}
        
        result.append({
            "ticketUid": t["ticketUid"],
            "flightNumber": t["flightNumber"],
            "fromAirport": f_data.get("fromAirport", "Unknown"),
            "toAirport": f_data.get("toAirport", "Unknown"),
            "date": f_data.get("date", "Unknown"),
            "status": t["status"],
            "price": t["price"]
        })
    return result

@app.get("/api/v1/me")
async def get_user_info(x_user_name: str = Header(...)):
    # Параллельно или последовательно собираем данные через клиентов
    tickets = await get_user_tickets(x_user_name)
    p_resp = await bonus_client.get_privilege(x_user_name)
    
    return {
        "tickets": tickets,
        "privilege": p_resp.json() if p_resp.status_code == 200 else {"balance": 0, "status": "BRONZE"}
    }

@app.post("/api/v1/tickets")
async def buy_ticket(request: dict, x_user_name: str = Header(...)):
    # 1. Проверка существования рейса
    f_resp = await flight_client.get_flight(request['flightNumber'])
    if f_resp.status_code != 200:
        raise HTTPException(status_code=400, detail="Flight not found")

    # 2. Бонусная операция
    bonus_payload = {
        "ticketUid": "00000000-0000-0000-0000-000000000000", # Временный ID для расчета
        "price": request['price'],
        "paidFromBalance": request['paidFromBalance']
    }
    b_resp = await bonus_client.calculate(x_user_name, bonus_payload)
    b_data = b_resp.json()

    # 3. Создание билета
    t_payload = {"flightNumber": request['flightNumber'], "price": request['price']}
    t_resp = await ticket_client.create_ticket(x_user_name, t_payload)
    t_data = t_resp.json()

    return {
        **t_data,
        "paidByMoney": request['price'] - b_data['paidByBonuses'],
        "paidByBonuses": b_data['paidByBonuses'],
        "privilege": b_data['privilege']
    }

@app.delete("/api/v1/tickets/{ticketUid}", status_code=204)
async def refund_ticket(ticketUid: str, x_user_name: str = Header(...)):
    # 1. Отмена в сервисе билетов
    t_resp = await ticket_client.delete_ticket(x_user_name, ticketUid)
    if t_resp.status_code == 404:
        raise HTTPException(status_code=404, detail="Ticket not found")
    
    # 2. Откат бонусов
    await bonus_client.rollback(x_user_name, ticketUid)
    return None


@app.get("/api/v1/tickets/{ticketUid}")
async def get_ticket_info(ticketUid: str, x_user_name: str = Header(...)):
    # 1. Запрашиваем данные о билете в Ticket Service
    t_resp = await ticket_client.get_ticket_by_uid(x_user_name, ticketUid)
    
    if t_resp.status_code == 404:
        raise HTTPException(status_code=404, detail="Билет не найден или не принадлежит пользователю")
    
    ticket = t_resp.json()
    print("="*30, ticket)

    # 2. Запрашиваем детали рейса во Flight Service по номеру рейса из билета
    f_resp = await flight_client.get_flight(ticket["flightNumber"])
    
    if f_resp.status_code != 200:
        # Если рейс не найден, возвращаем базовую информацию из билета
        return {
            **ticket,
            "fromAirport": "Unknown",
            "toAirport": "Unknown",
            "date": "Unknown"
        }

    flight_data = f_resp.json()

    # 3. Формируем финальный ответ согласно спецификации TicketResponse
    return {
        "ticketUid": ticket["ticketUid"],
        "flightNumber": ticket["flightNumber"],
        "fromAirport": flight_data["fromAirport"],
        "toAirport": flight_data["toAirport"],
        "date": flight_data["date"],
        "status": ticket["status"],
        "price": ticket["price"]
    }

