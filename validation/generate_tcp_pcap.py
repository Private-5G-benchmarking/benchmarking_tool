import socket

def client_program():
    host = '172.30.0.47'  
    port = 9999 

    client_socket = socket.socket() 
    client_socket.connect((host, port))  

    for i in range(10000):
        message = f"This is packet {i}"
        client_socket.send(message.encode()) 
        data = client_socket.recv(1024).decode() 


    client_socket.close()  # close the connection


if __name__ == '__main__':
    client_program()
