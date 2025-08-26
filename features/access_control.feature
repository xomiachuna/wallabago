Feature: Access Control
    Background:
        Given there exists a user account
        And there exists a client
    
    Scenario Outline: Client Authentication
        Given client credentials are <client>
        And user credentials are <user>
        When client uses credentials to authenticate
        Then the client should be <result>

        Examples:
            |client |user   |result         |
            |valid  |valid  |authenticated  |
            |invalid|valid  |rejected       |
            |valid  |invalid|rejected       |
            |invalid|invalid|rejected       |
