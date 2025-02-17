package bookingclient;

import java.util.UUID;

public class UnmarshalledPacket {
    private final UUID originalMessageID;
    private final int status;
    private final byte[] payload;

    public UnmarshalledPacket(UUID originalMessageID, int status, byte[] payload) {
        this.originalMessageID = originalMessageID;
        this.status = status;
        this.payload = payload;
    }
    public UUID getOriginalMessageId() {
        return originalMessageID;
    }

    public int getStatus() {
        return status;
    }

    public byte[] getPayload() {
        return payload;
    }
}
