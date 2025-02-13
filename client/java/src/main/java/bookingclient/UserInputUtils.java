package bookingclient;

import java.util.Scanner;

public class UserInputUtils {
    /**
     * A static `Scanner` instance initialized with `System.in`, used to read user input from the console.
     * This instance can be reused throughout the class to avoid creating multiple scanner objects,
     * reducing resource usage and ensuring consistent input handling.
     */
    private final static Scanner scanner = new Scanner(System.in);

    // Method to get a valid String input from the user
    public static String getStringInput(String prompt) {
        System.out.println(prompt); // Display the prompt to the user
        return scanner.nextLine().trim(); // Read and trim any surrounding spaces from the input
    }

    // Method to get a valid Integer input from the user
    public static int getIntInput(String prompt) {
        System.out.println(prompt); // Display the prompt to the user
        int number = scanner.nextInt();
        scanner.nextLine();
        return number;
    }
}
