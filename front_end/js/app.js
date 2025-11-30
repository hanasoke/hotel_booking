const API_BASE_URL = 'http://localhost:8080/api';

// Load rooms when page loads
document.addEventListener('DOMContentLoaded', function() {
    loadRooms();
    setupDateListeners();
});

async function loadRooms() {
    try {
        const response = await fetch(`${API_BASE_URL}/rooms`);
        const result = await response.json();
        
        if (result.data) {
            displayRooms(result.data);
        }
    } catch (error) {
        console.error('Error loading rooms:', error);
        alert('Gagal memuat data kamar');
    }
}

function displayRooms(rooms) {
    const container = document.getElementById('rooms-container');
    container.innerHTML = '';

    rooms.forEach(room => {
        const roomCard = `
            <div class="col-md-6 col-lg-3">
                <div class="card room-card">
                    <img src="${room.image_url || '/images/default-room.jpg'}" class="card-img-top" alt="${room.type}">
                    <div class="card-body">
                        <h5 class="card-title">Kamar ${room.room_number}</h5>
                        <h6 class="card-subtitle mb-2 text-muted">${room.type}</h6>
                        <p class="card-text">${room.description}</p>
                        <p class="card-text">
                            <i class="fas fa-user"></i> ${room.capacity} orang<br>
                            <strong>Rp ${room.price.toLocaleString()}/malam</strong>
                        </p>
                        <div class="d-grid">
                            <button class="btn btn-primary" onclick="openBookingModal(${room.id}, ${room.price})" 
                                ${!room.is_available ? 'disabled' : ''}>
                                ${room.is_available ? 'Pesan Sekarang' : 'Tidak Tersedia'}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `;
        container.innerHTML += roomCard;
    });
}

function openBookingModal(roomId, price) {
    document.getElementById('roomId').value = roomId;
    document.getElementById('totalAmount').value = 'Rp 0';
    
    // Set min date to today
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('checkInDate').min = today;
    document.getElementById('checkOutDate').min = today;
    
    const modal = new bootstrap.Modal(document.getElementById('bookingModal'));
    modal.show();
}

function setupDateListeners() {
    document.getElementById('checkInDate').addEventListener('change', calculateTotal);
    document.getElementById('checkOutDate').addEventListener('change', calculateTotal);
}

function calculateTotal() {
    const checkIn = document.getElementById('checkInDate').value;
    const checkOut = document.getElementById('checkOutDate').value;
    const roomId = document.getElementById('roomId').value;

    if (checkIn && checkOut && roomId) {
        const checkInDate = new Date(checkIn);
        const checkOutDate = new Date(checkOut);
        
        if (checkOutDate <= checkInDate) {
            alert('Tanggal check-out harus setelah tanggal check-in');
            document.getElementById('totalAmount').value = 'Rp 0';
            return;
        }

        // In a real app, we would fetch the room price from API
        // For now, we'll use a fixed price (you can modify this)
        const days = Math.ceil((checkOutDate - checkInDate) / (1000 * 60 * 60 * 24));
        const pricePerNight = 250000; // This should come from room data
        const total = days * pricePerNight;
        
        document.getElementById('totalAmount').value = `Rp ${total.toLocaleString()}`;
    }
}

async function submitBooking() {
    const formData = {
        room_id: parseInt(document.getElementById('roomId').value),
        customer_name: document.getElementById('customerName').value,
        customer_email: document.getElementById('customerEmail').value,
        customer_phone: document.getElementById('customerPhone').value,
        check_in_date: document.getElementById('checkInDate').value,
        check_out_date: document.getElementById('checkOutDate').value
    };

    // Validation
    if (!formData.customer_name || !formData.customer_email || !formData.customer_phone || 
        !formData.check_in_date || !formData.check_out_date) {
        alert('Harap lengkapi semua field');
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/bookings`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(formData)
        });

        if (response.ok) {
            const result = await response.json();
            alert('Booking berhasil dibuat! ID Booking: ' + result.data.id);
            
            // Close modal and reset form
            const modal = bootstrap.Modal.getInstance(document.getElementById('bookingModal'));
            modal.hide();
            document.getElementById('bookingForm').reset();
            
            // Reload rooms
            loadRooms();
        } else {
            const error = await response.json();
            alert('Gagal membuat booking: ' + error.error);
        }
    } catch (error) {
        console.error('Error creating booking:', error);
        alert('Terjadi kesalahan saat membuat booking');
    }
}