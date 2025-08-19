Feature: OAuth
    Background:
        Given the following users exist:
            | username | password | isAdmin |
            | admin | admin | true |
            | user | user | false |
        And the following clients exist:
            | client-id | client-secret |
            | web | web |
            | alternate | alternate |

    Scenario Outline: Valid Client Credentials Flow
        Use valid credentials

        Given client id is "<client-id>"
        And client secret is "<client-secret>"
        And username is "<username>"
        And password "<password>"

        When token is requested with client credentials flow

        Then refresh token should be returned
        And no error should be returned
        # And refresh token can be used to obtain a new token

        Examples:
            |client-id|client-secret|username|password|
            |web|web|admin|admin|
            |web|web|user|user|
            |alternate|alternate|admin|admin|
            |alternate|alternate|user|user|

            #     Scenario Outline: Invalid Client Credentials Flow
            #         Use bad credentials
            # 
            #         Given client id is <client-id> 
            #         And client secret is <client-secret> 
            #         And username is <username>
            #         And password <password>
            # 
            #         When a token is requested with client credentials flow
            # 
            #         Then an error is returned
            #         And no token is returned
            #         And no refresh token is returned
            # 
            #         Examples:
            #             |client-id|client-secret|username|password|
            #             |web|web|admin|admix|
            #             |web|wex|user|user|
            #             |alternat|alternate|admin|admin|
            #             |alternate|alternate|user|user|
