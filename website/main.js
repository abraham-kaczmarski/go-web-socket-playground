const store = {
  apiUrl: "127.0.0.1:5000",
  html: {
    activeRoomId: document.getElementById("active-room-id"),
    roomCreateForm: document.getElementById("room-create-form"),
    roomCreateFormStatus: document.getElementById("room-create-form-status"),
    roomList: document.getElementById("room-list"),
    roomsSection: document.getElementById("rooms-section"),
  },
  room: {
    id: null,
    socket: null,
  },
}

main()

function main() {
  const { roomCreateForm } = store.html

  loadRooms()

  roomCreateForm.addEventListener("submit", event => {
    event.preventDefault()

    const formData = new FormData(event.target)
    const data = Object.fromEntries(formData.entries())

    createRoom(data)
  })
}

async function loadRooms() {
  const { apiUrl } = store
  const { roomList } = store.html

  roomList.innerText = "Loading..."

  try {
    const res = await fetch(`http://${apiUrl}/rooms`)

    if (res.status !== 200) {
      throw new Error()
    }

    const rooms = await res.json()

    if (rooms.length === 0) {
      roomList.innerText = "There are no active rooms."
      return
    }

    roomList.innerHTML = ""

    for (const room of rooms) {
      const listElement = createRoomListElement(room)
      roomList.appendChild(listElement)
    }
  } catch {
    console.log(err)
    roomList.innerText = "Loading rooms failed."
  }
}

async function createRoom(data) {
  const { apiUrl } = store
  const { roomCreateFormStatus } = store.html

  try {
    const res = await fetch(`http://${apiUrl}/rooms`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    })

    if (res.status !== 201) {
      throw new Error()
    }

    loadRooms()
  } catch (err) {
    console.log(err)
    roomCreateFormStatus.innerText = "Creating room failed."
  }
}

function createRoomListElement(room) {
  const listElement = document.createElement("li")

  listElement.innerText = room.name

  const button = document.createElement("button")
  button.innerText = "Join"
  button.addEventListener("click", () => joinRoom(room.id))

  listElement.appendChild(button)

  return listElement
}

function joinRoom(roomId) {
  store.room.id = roomId
  store.html.activeRoomId.innerHTML = roomId
  openSocket()
}

function openSocket() {
  const { apiUrl } = store

  const socket = new WebSocket(`ws://${apiUrl}/rooms/ws`)

  socket.onopen = () => {
    socket.send({ data: "Yello from the client!" })
  }

  socket.onclose = () => {
    socket.send("Until next time!")
  }

  socket.onmessage = event => {
    console.log(event.data)
  }

  socket.onerror = error => {
    console.log("Socker error:", error)
  }

  store.room.socket = socket
}
