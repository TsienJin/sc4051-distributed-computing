package bookingclient;
import java.io.IOException;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.UnknownHostException;
import java.nio.ByteBuffer;
import java.util.Arrays;

public class NetworkHandler {
    private static final String HOST = "100.105.193.66";
    private static final int UDP_PORT = 8765;
    private static final int TIMEOUT_MS = 2000;
    private static final int MAX_RETRIES = 1;
    private static DatagramSocket socket;
    private static InetAddress address;

    public void networkClient(){
        try {
            socket = new DatagramSocket();
            address = InetAddress.getByName("100.105.193.66");
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public byte[] sendPacketWithAck(byte[] packet) throws IOException {
        //This function sends a packet and waits for an acknowledgement and response.
        DatagramPacket DatagramPacket = new DatagramPacket(packet, packet.length, address, UDP_PORT);
        byte[] ackBuffer = new byte[1024];
        DatagramPacket ackPacket = new DatagramPacket(ackBuffer, ackBuffer.length);
        int retries = 0;
        while (retries < MAX_RETRIES) {
            try {
                System.out.println(bytesToHex(DatagramPacket.getData()));
                socket.send(DatagramPacket);
                socket.receive(ackPacket);
                System.out.println("ack time" + bytesToHex(ackPacket.getData()));
                socket.receive(ackPacket);
                System.out.println("second time" + bytesToHex(ackPacket.getData()));
                return ackPacket.getData();
            } catch (IOException e) {
                retries++;
                System.out.println("Failed to send packet, retrying...");
            }
        }

        throw new IOException("Failed to send packet after " + MAX_RETRIES + " retry attempts.");

    }
    public static boolean isAck(byte[] packet){
        //Helper method to check if a packet recieved is a response
        if(packet.length < 18){
            throw new IllegalArgumentException("Packet too short");
        }else {
            byte packetTypeBytes = packet[18];
            if(packetTypeBytes != PacketType.ACK.getCode()){
                return false;
            }
            return true;
        }
    }
    // Helper method to convert byte array to hex string for easier debugging
    public static String bytesToHex(byte[] byteArray) {
        StringBuilder hexString = new StringBuilder();
        for (byte b : byteArray) {
            hexString.append(String.format("%02X", b));
        }
        return hexString.toString();
    }
    private void backoffDelay(int retryCount) {
        //Helper function to introduce backoffDelay to simulate server down time
        try {
            // Base delay: 100ms multiplied by 2^retryCount.
            long baseDelay = 100 * (long) Math.pow(2, retryCount);

            // Random factor between 0.8 and 1.2.
            double randomFactor = 0.8 + Math.random() * 0.4;

            // Apply the random factor to the base delay.
            long delay = (long) (baseDelay * randomFactor);
            Debugger.log("Delaying for " + delay + "ms before retrying...");
            Thread.sleep(delay);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }


}
