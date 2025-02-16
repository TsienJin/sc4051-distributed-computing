package bookingclient;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.time.LocalTime;

interface ClientState {
    void handleRequest(Client client);
}
class MenuState implements ClientState{
    //1,2,3,4 are MANDATORY. 5,6,7,8 not MANDATORY
    @Override
    public void handleRequest(Client client) {
        System.out.println("1. Query Facility Availability");
        System.out.println("2. Book Facility");
        System.out.println("3. Modify Booking");
        System.out.println("4. Monitor Facility");
        System.out.println("5. Create Facility");
        System.out.println("6. Delete Facility");
        System.out.println("7. List Facilities");
        System.out.println("8. Delete Booking");
        System.out.println("9. Exit");

        int choice = UserInputUtils.getIntInput("Choose an option:");
        switch (choice) {
            case 1:
                client.setState(new QueryFacilityState());
                break;
            case 2:
                client.setState(new BookFacilityState());
                break;
            case 5:
                client.setState(new CreateFacilityState());
                break;
            case 6:
                client.setState(new DeleteFacilityState());
                break;
            case 8:
                client.setState(new DeleteBookingState());
                break;
            case 9:
                System.out.println("Exiting system...");
                System.exit(0);
            default:
                System.out.println("Invalid option. Please try again.");
        }
        client.handleRequest();
    }

}

class CreateFacilityState implements ClientState {
    @Override
    public void handleRequest(Client client) {
        // Get user input for facility name
        String facility = UserInputUtils.getStringInput("Create Facility Name:");
        System.out.println("Creating Facility Name: " + facility);

        // Create PacketMarshaller and NetworkHandler objects (no singleton here, just direct instantiation)
        PacketMarshaller marshaller = new PacketMarshaller();  // Direct instantiation
        byte[] packet = marshaller.marshalCreateFacilityRequest(facility);  // Marshal the facility data
        byte[] ackpacket = null;
        // Directly create the NetworkHandler and send the packet
        NetworkHandler networkHandler = new NetworkHandler();  // Direct instantiation
        networkHandler.networkClient();
        try {
            ackpacket = networkHandler.sendPacketWithAck(packet);  // Send packet with acknowledgment handling
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
        marshaller.unmarshalResponse(ackpacket);
        // After processing, return to MenuState
        client.setState(new MenuState());
        client.handleRequest();
    }
}

class DeleteFacilityState implements ClientState{
    @Override
    public void handleRequest(Client client) {
        String facility = UserInputUtils.getStringInput("Delete Facility Name:");
        System.out.println("Deleting Facility Name " + facility);
        // Create PacketMarshaller and NetworkHandler objects (no singleton here, just direct instantiation)
        PacketMarshaller marshaller = new PacketMarshaller();  // Direct instantiation
        byte[] packet = marshaller.marshalDeleteFacilityRequest(facility);  // Marshal the facility data
        byte[] ackpacket = null;
        // Directly create the NetworkHandler and send the packet
        NetworkHandler networkHandler = new NetworkHandler();  // Direct instantiation
        networkHandler.networkClient();
        try {
            ackpacket = networkHandler.sendPacketWithAck(packet);  // Send packet with acknowledgment handling
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
        marshaller.unmarshalResponse(ackpacket);
        // After processing, return to MenuState
        client.setState(new MenuState());
        client.handleRequest();
    }
}

class QueryFacilityState implements ClientState{
    @Override
    public void handleRequest(Client client) {
        String facility = UserInputUtils.getStringInput("Query Facility Name:");
        int numberOfDays = UserInputUtils.getIntInput("Number of Days:");

        System.out.println("Querying Facility Name " + facility);
        // Create PacketMarshaller and NetworkHandler objects (no singleton here, just direct instantiation)
        PacketMarshaller marshaller = new PacketMarshaller();  // Direct instantiation
        byte[] packet = marshaller.marshalQueryFacilityRequest(facility, numberOfDays);  // Marshal the facility data
        byte[] ackpacket = null;
        // Directly create the NetworkHandler and send the packet
        NetworkHandler networkHandler = new NetworkHandler();  // Direct instantiation
        networkHandler.networkClient();
        try {
            ackpacket = networkHandler.sendPacketWithAck(packet);  // Send packet with acknowledgment handling
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
        marshaller.unmarshalResponse(ackpacket);
        // After processing, return to MenuState
        client.setState(new MenuState());
        client.handleRequest();

        client.setState(new MenuState());
        client.handleRequest();
    }
}

class BookFacilityState implements ClientState{
    @Override
    public void handleRequest(Client client) {
        String facility = UserInputUtils.getStringInput("Query Facility Name:");
        int startTime = UserInputUtils.getIntInput("Enter hours since UNIX (startTime: " + getHoursSinceUnix() + "): ");
        int endTime = UserInputUtils.getIntInput("Enter hours since UNIX (endTime: " + getHoursSinceUnix() + "): ");

        System.out.println("Booking Facility Name " + facility);
        // Create PacketMarshaller and NetworkHandler objects (no singleton here, just direct instantiation)
        PacketMarshaller marshaller = new PacketMarshaller();  // Direct instantiation
        byte[] packet = marshaller.marshalBookFacilityRequest(facility, startTime, endTime);  // Marshal the facility data
        byte[] ackpacket = null;
        // Directly create the NetworkHandler and send the packet
        NetworkHandler networkHandler = new NetworkHandler();  // Direct instantiation
        networkHandler.networkClient();
        try {
            ackpacket = networkHandler.sendPacketWithAck(packet);  // Send packet with acknowledgment handling
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
        marshaller.unmarshalResponse(ackpacket);
        // After processing, return to MenuState
        client.setState(new MenuState());
        client.handleRequest();

        client.setState(new MenuState());
        client.handleRequest();
    }

    public long getHoursSinceUnix() {
        // Get the current time in milliseconds since Unix epoch
        long currentTimeMillis = System.currentTimeMillis();

        // Convert milliseconds to seconds
        long currentTimeSeconds = currentTimeMillis / 1000;

        // Convert seconds to hours
        long hoursSinceUnixEpoch = currentTimeSeconds / 3600;  // 3600 seconds in an hour

        return hoursSinceUnixEpoch;
    }
}

class DeleteBookingState implements ClientState {
    @Override
    public void handleRequest(Client client) {
        int confirmationCode = UserInputUtils.getHexInput("Delete Booking for Confirmation code:");

        System.out.println("Deleting booking with confirmation code: " + confirmationCode);
        // Create PacketMarshaller and NetworkHandler objects (no singleton here, just direct instantiation)
        PacketMarshaller marshaller = new PacketMarshaller();  // Direct instantiation
        byte[] packet = marshaller.marshalDeleteBookingRequest(confirmationCode);  // Marshal the facility data
        byte[] ackpacket = null;
        // Directly create the NetworkHandler and send the packet
        NetworkHandler networkHandler = new NetworkHandler();  // Direct instantiation
        networkHandler.networkClient();
        try {
            ackpacket = networkHandler.sendPacketWithAck(packet);  // Send packet with acknowledgment handling
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
        marshaller.unmarshalResponse(ackpacket);
        // After processing, return to MenuState
        client.setState(new MenuState());
        client.handleRequest();

        client.setState(new MenuState());
        client.handleRequest();
    }
}


class Client {
    private ClientState currentState;

    public Client() {
        this.currentState = new MenuState();  // Start in the Menu state
    }

    public void setState(ClientState state) {
        this.currentState = state;  // Change the current state
    }

    public void handleRequest() {
        currentState.handleRequest(this);  // Delegate to the current state
    }
}
